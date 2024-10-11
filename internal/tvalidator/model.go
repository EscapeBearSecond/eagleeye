package tvalidator

type dbTemplate struct {
	TemplateID      string `db:"template_id"`
	TemplateContent string `db:"template_content"`
}
