package tvalidator

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"codeup.aliyun.com/66d825f8c06a2fdac7bbfe8c/eagleeye/internal/util"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/projectdiscovery/nuclei/v3/pkg/catalog/disk"
	"github.com/projectdiscovery/nuclei/v3/pkg/templates"

	"github.com/samber/lo"
)

func ExecuteDbCommand(driver, dsn, sql string) error {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	dbTemplates := []*dbTemplate{}
	err = db.Select(&dbTemplates, sql)
	if err != nil {
		return fmt.Errorf("unable to get templates: %w", err)
	}

	if len(dbTemplates) == 0 {
		return errors.New("no templates found, please check your sql: template_id, template_content are required")
	}

	templatesMap := lo.SliceToMap(dbTemplates, func(t *dbTemplate) (string, string) {
		return fmt.Sprintf("%s.yaml", t.TemplateID), t.TemplateContent
	})

	errs := make([]error, 0)

	parser := templates.NewParser()
	catalog := newCatalog(templatesMap)
	tagFilter, _ := templates.NewTagFilter(&templates.TagFilterConfig{})

	for templateID := range templatesMap {
		_, err = parser.LoadTemplate(templateID, tagFilter, nil, catalog)
		if err != nil {
			errs = append(errs, fmt.Errorf("template [%s] parse failed: %w", templateID, err))
		}
	}

	if len(errs) > 0 {
		return util.JoinErrors(errs, &util.ErrorsOptions{UseIndex: true, Separator: "\n"})
	}

	return nil
}

func ExecuteFsCommand(dir string) error {
	errs := make([]error, 0)

	catalog := disk.NewCatalog(dir)
	parser := templates.NewParser()
	tagFilter, _ := templates.NewTagFilter(&templates.TagFilterConfig{})

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk dir failed: %w", err)
		}

		if d.IsDir() {
			return nil
		}

		_, err = parser.LoadTemplate(path, tagFilter, nil, catalog)
		if err != nil {
			errs = append(errs, fmt.Errorf("template [%s] parse failed: %w", path, err))
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(errs) > 0 {
		return util.JoinErrors(errs, &util.ErrorsOptions{UseIndex: true, Separator: "\n"})
	}
	return nil
}
