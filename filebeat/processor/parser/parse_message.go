package parser

import (
	"regexp"
	"strings"
	"unsafe"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/processors"
	"github.com/pkg/errors"
)

type parseMessage struct {
	regex *regexp.Regexp
}

func init() {
	processors.RegisterPlugin("parse_message", newParseMessage)
}

func newParseMessage(c *common.Config) (processors.Processor, error) {
	regex, err := regexp.Compile("(?P<timedate>^\\d{4}-\\d{2}-\\d{2} \\d{1,2}:\\d{1,2}:\\d{1,2}(\\.\\d+)*)\\s+(?P<log_level>[Aa]lert|ALERT|[Tt]race|TRACE|[Dd]ebug|DEBUG|[Nn]otice|NOTICE|[Ii]nfo|INFO|[Ww]arn(?:ing)?|WARN(?:ING)?|[Ee]rr(?:or)?|ERR(?:OR)?|[Cc]rit(?:ical)?|CRIT(?:ICAL)?|[Ff]atal|FATAL|[Ss]evere|SEVERE|[Ee]merg(?:ency)?|EMERG(?:ENCY))\\s+\\[(?P<ext_info>.*?)\\](?P<content>[\\s\\S]*$)")
	if err != nil {
		return nil, errors.Wrap(err, "parse failed")
	}

	return &parseMessage{regex: regex}, nil
}

func (p *parseMessage) Run(event *beat.Event) (*beat.Event, error) {
	message, err := event.GetValue("message")
	if err != nil {
		return event, errors.Wrap(err, "fail to get message value")
	}
	out, tags := p.parseV2(message.(string))
	event.PutValue("message", out)
	common.MergeFieldsDeep(event.Fields, common.MapStr{"terminus": common.MapStr{"tags": tags}}, true)
	return event, nil
}

func extractTags(raw string) (tagstr string, tags map[string]interface{}) {
	tagstr, tags = "", make(map[string]interface{})
	for idx, item := range strings.Split(raw, ",") {
		tmp := strings.Split(item, "=")
		if len(tmp) == 2 {
			tags[tmp[0]] = tmp[1]
			continue
		}

		switch idx {
		case 1:
			tags["request-id"] = item
		}

		if tagstr == "" {
			tagstr = item
		} else {
			tagstr += "," + item
		}
	}

	return
}

// https://yuque.antfin.com/spot/develop-docs/gkfxgl
func (p *parseMessage) parseV2(message string) (string, map[string]interface{}) {
	if !p.regex.Match(stringToBytes(message)) {
		return message, map[string]interface{}{}
	}

	groupNames := p.regex.SubexpNames()
	level, tagstr := "", ""
	for _, matches := range p.regex.FindAllStringSubmatch(message, -1) {
		for idx, name := range groupNames {
			switch name {
			case "log_level":
				level = matches[idx]
			case "ext_info":
				tagstr = matches[idx]
			}
		}
	}

	newTagstr, tags := extractTags(tagstr)
	tags["level"] = level
	return strings.ReplaceAll(message, tagstr, newTagstr), tags
}

func (p *parseMessage) parse(message string) (string, map[string]interface{}) {
	matches := p.regex.FindStringSubmatch(message)
	if len(matches) < 4 {
		return message, nil
	}

	m := make(map[string]interface{})
	m["level"] = strings.ToUpper(matches[2])
	replace := "$1"

	attrs := strings.Split(matches[3], ",")
	if len(attrs) < 3 {
		return message, m
	}
	m["request-id"] = attrs[1]
	replace += " - [" + strings.Join(attrs[0:3], ",") + "]"

	for i := 2; i < len(attrs); i++ {
		kv := strings.Split(attrs[i], ":")
		if len(kv) == 2 {
			m[kv[0]] = kv[1]
		}
	}
	return p.regex.ReplaceAllString(message, replace), m
}

func (*parseMessage) String() string {
	return "parse_message"
}

func stringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}
