package engine

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// IniValue is one key=value occurrence together with the comment block that
// preceded it. A nil Comment means "no comment" and keeps any existing comment
// when a value is overwritten (Source/utils/ini.hpp:22).
type IniValue struct {
	Comment *string
	Value   string
}

// iniEntry holds every value of one key plus its insertion index; upstream
// keeps the same explicit index for document order (Source/utils/ini.hpp:147).
type iniEntry struct {
	index  int
	key    string
	values []IniValue
}

// iniSection holds one [section] block: the comment block before its header,
// its name, insertion index, and keys in insertion order
// (Source/utils/ini.hpp:147).
type iniSection struct {
	index   int
	name    string
	comment string
	entries []*iniEntry
	byKey   map[string]*iniEntry
}

// Ini is a parsed diablo.ini document (Source/utils/ini.cpp). Sections and
// keys keep document order via indexes so serialize is stable.
type Ini struct {
	sections []*iniSection
	byName   map[string]*iniSection
	changed  bool
}

// iniLine is one raw line plus where it ends, mirroring the lineBegin/lineEnd
// window Ini::parse walks (Source/utils/ini.cpp:137).
type iniLine struct {
	text    string
	lineNum int
}

// iniCursor carries the parse state between lines: the pending comment span,
// the current section, and any fatal error.
type iniCursor struct {
	ini        *Ini
	current    *iniSection
	comment    string
	hasComment bool
}

// ParseIni parses diablo.ini bytes. It is a 1:1 port of Ini::parse
// (Source/utils/ini.cpp:137): the same whitespace rules, the same UTF-8 BOM
// skip, the same errors, and the same comment-block ownership.
func ParseIni(data []byte) (*Ini, error) {
	text := strings.TrimPrefix(string(data), "\xef\xbb\xbf") // SkipUtf8Bom (ini.cpp:63)

	cursor := &iniCursor{ini: &Ini{byName: map[string]*iniSection{}}}
	for line := range iniLines(text) {
		if err := cursor.parseLine(line); err != nil {
			return nil, err
		}
	}
	return cursor.ini, nil
}

// iniLines splits the document into lines, dropping a trailing \r before \n
// like the lineEnd adjustment (Source/utils/ini.cpp:145-150).
func iniLines(text string) func(yield func(iniLine) bool) {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(text, "\n")
	return func(yield func(iniLine) bool) {
		for i, raw := range lines {
			if !yield(iniLine{text: raw, lineNum: i + 1}) {
				return
			}
		}
	}
}

// parseLine classifies one physical line: blank, comment, section header, or
// key=value (Source/utils/ini.cpp:151-205).
func (c *iniCursor) parseLine(line iniLine) error {
	trimmed := strings.TrimLeft(line.text, " \t")
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, ";") {
		c.noteComment(line.text)
		return nil
	}
	comment := c.takeComment()
	if strings.HasPrefix(trimmed, "[") {
		return c.parseSection(trimmed, comment, line.lineNum)
	}
	return c.parseKeyValue(trimmed, comment, line.lineNum)
}

// noteComment appends to the pending comment span; the span runs to the start
// of the next non-comment line, trailing newline included
// (std::string_view(commentBegin, lineBegin - commentBegin), ini.cpp:160).
func (c *iniCursor) noteComment(text string) {
	if !c.hasComment {
		c.comment, c.hasComment = text+"\n", true
		return
	}
	c.comment += text + "\n"
}

func (c *iniCursor) takeComment() string {
	if !c.hasComment {
		return ""
	}
	c.hasComment = false
	return c.comment
}

// parseSection handles [name] headers: new sections get the pending comment
// and the next insertion index; repeats re-point the cursor
// (Source/utils/ini.cpp:171-194).
func (c *iniCursor) parseSection(trimmed, comment string, lineNum int) error {
	body := trimmed[1:]
	end := strings.IndexByte(body, ']')
	if end == -1 {
		return fmt.Errorf("line %d: unclosed section name %s", lineNum, body)
	}
	name := body[:end]
	after := strings.TrimSpace(body[end+1:])
	if after != "" {
		return fmt.Errorf("line %d: content after section [%s]: %s", lineNum, name, after)
	}
	section := c.ini.byName[name]
	if section == nil {
		section = &iniSection{
			index:   len(c.ini.sections),
			name:    name,
			comment: comment,
			byKey:   map[string]*iniEntry{},
		}
		c.ini.sections = append(c.ini.sections, section)
		c.ini.byName[name] = section
	}
	c.current = section
	return nil
}

// parseKeyValue handles key=value lines: leading whitespace is trimmed from
// the value, repeated keys append (Source/utils/ini.cpp:196-205).
func (c *iniCursor) parseKeyValue(trimmed, comment string, lineNum int) error {
	if c.current == nil {
		return fmt.Errorf("line %d: key not in any section", lineNum)
	}
	eq := strings.IndexByte(trimmed, '=')
	if eq == -1 {
		return fmt.Errorf("line %d: key %s has no value", lineNum, trimmed)
	}
	key := strings.TrimSpace(trimmed[:eq])
	value := strings.TrimLeft(trimmed[eq+1:], " ")
	v := IniValue{Value: value}
	if comment != "" {
		stored := comment
		v.Comment = &stored
	}
	entry := c.current.byKey[key]
	if entry == nil {
		entry = &iniEntry{index: len(c.current.entries), key: key}
		c.current.entries = append(c.current.entries, entry)
		c.current.byKey[key] = entry
	}
	entry.values = append(entry.values, v)
	return nil
}

// Get returns every value of the key in document order, or nil
// (Source/utils/ini.cpp:215).
func (in *Ini) Get(section, key string) []IniValue {
	entry := in.entry(section, key)
	if entry == nil {
		return nil
	}
	return entry.values
}

// GetString returns the last value, honoring the empty-value default rule
// (Source/utils/ini.cpp:226).
func (in *Ini) GetString(section, key, defaultValue string) string {
	xs := in.Get(section, key)
	if len(xs) == 0 || xs[len(xs)-1].Value == "" {
		return defaultValue
	}
	return xs[len(xs)-1].Value
}

// GetInt parses the last value as int; upstream aborts the app on garbage
// (Source/utils/ini.cpp:232), here the error is returned instead.
func (in *Ini) GetInt(section, key string, defaultValue int) (int, error) {
	raw, ok := in.lastValue(section, key)
	if !ok {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue, fmt.Errorf("ini: failed to parse %s.%s=%s as int", section, key, raw)
	}
	return v, nil
}

// GetBool accepts exactly "0"/"1" like upstream (Source/utils/ini.cpp:246).
func (in *Ini) GetBool(section, key string, defaultValue bool) (bool, error) {
	raw, ok := in.lastValue(section, key)
	if !ok {
		return defaultValue, nil
	}
	switch raw {
	case "0":
		return false, nil
	case "1":
		return true, nil
	}
	return defaultValue, fmt.Errorf("ini: failed to parse %s.%s=%s as bool", section, key, raw)
}

// GetFloat parses the last value as float32 like upstream (ini.cpp:258).
func (in *Ini) GetFloat(section, key string, defaultValue float64) (float64, error) {
	raw, ok := in.lastValue(section, key)
	if !ok {
		return defaultValue, nil
	}
	v, err := strconv.ParseFloat(raw, 32)
	if err != nil {
		return defaultValue, fmt.Errorf("ini: failed to parse %s.%s=%s as float", section, key, raw)
	}
	return v, nil
}

// SetString sets a single new value (Source/utils/ini.cpp:354).
func (in *Ini) SetString(section, key, value string) {
	in.setValues(section, key, []IniValue{{Value: value}})
}

// SetInt stores the decimal form like StrCat(int) upstream (ini.cpp:365).
func (in *Ini) SetInt(section, key string, value int) {
	in.setValues(section, key, []IniValue{{Value: strconv.Itoa(value)}})
}

// SetBool stores "1"/"0" upstream style (ini.cpp:370).
//
//nolint:revive // value is upstream's setBool signature, not a control flag.
func (in *Ini) SetBool(section, key string, value bool) {
	stored := "0"
	if value {
		stored = "1"
	}
	in.setValues(section, key, []IniValue{{Value: stored}})
}

// SetFloat stores the shortest round-trip form of the float32 value like
// std::to_chars (ini.cpp:375).
func (in *Ini) SetFloat(section, key string, value float64) {
	in.setValues(section, key, []IniValue{{Value: strconv.FormatFloat(value, 'g', -1, 32)}})
}

// SetStrings stores one value or an explicit value list (ini.cpp:337).
func (in *Ini) SetStrings(section, key string, values []string) {
	parsed := make([]IniValue, len(values))
	for i, s := range values {
		parsed[i] = IniValue{Value: s}
	}
	in.setValues(section, key, parsed)
}

// Serialize renders the document: sections in insertion order separated by a
// blank line, keys in insertion order, comments verbatim with \n normalized to
// \r\n (Source/utils/ini.cpp:415).
func (in *Ini) Serialize() string {
	sections := slices.Clone(in.sections)
	slices.SortStableFunc(sections, func(a, b *iniSection) int { return a.index - b.index })

	parts := make([]string, 0, len(sections))
	for _, sec := range sections {
		parts = append(parts, serializeSection(sec))
	}
	return strings.Join(parts, "\r\n")
}

// serializeSection renders one [section] block without the inter-section
// separator.
func serializeSection(sec *iniSection) string {
	lines := make([]string, 0, 2*len(sec.entries)+2)
	if sec.comment != "" {
		lines = append(lines, commentLines(sec.comment)...)
	}
	lines = append(lines, "["+sec.name+"]\r\n")
	entries := slices.Clone(sec.entries)
	slices.SortStableFunc(entries, func(a, b *iniEntry) int { return a.index - b.index })
	for _, entry := range entries {
		lines = append(lines, serializeEntry(entry)...)
	}
	return strings.Join(lines, "")
}

// serializeEntry renders every value of one key with its comment block.
func serializeEntry(entry *iniEntry) []string {
	lines := make([]string, 0, 2*len(entry.values))
	for _, v := range entry.values {
		if v.Comment != nil && *v.Comment != "" {
			lines = append(lines, commentLines(*v.Comment)...)
		}
		lines = append(lines, entry.key+"="+v.Value+"\r\n")
	}
	return lines
}

// commentLines normalizes lone \n to \r\n like upstream AppendComment
// (ini.cpp:389).
func commentLines(comment string) []string {
	text := strings.ReplaceAll(strings.ReplaceAll(comment, "\r\n", "\n"), "\n", "\r\n")
	out := make([]string, 0, strings.Count(text, "\n")+1)
	for _, line := range strings.SplitAfter(text, "\n") {
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// equalValues compares only the value strings; comments are metadata.
func equalValues(a, b []IniValue) bool {
	for i := range a {
		if a[i].Value != b[i].Value {
			return false
		}
	}
	return true
}

// lastValue returns the final value string of the key, or "" when absent.
func (in *Ini) lastValue(section, key string) (string, bool) {
	xs := in.Get(section, key)
	if len(xs) == 0 || xs[len(xs)-1].Value == "" {
		return "", false
	}
	return xs[len(xs)-1].Value, true
}

// entry resolves section+key to the entry, or nil.
func (in *Ini) entry(section, key string) *iniEntry {
	sec := in.byName[section]
	if sec == nil {
		return nil
	}
	return sec.byKey[key]
}

// setValues mirrors Ini::set(Values&&) (Source/utils/ini.cpp:272): create with
// append-order index, delete on empty, no-op when equal, comment preservation
// where the new value has none (ini.cpp:333-336).
func (in *Ini) setValues(section, key string, values []IniValue) {
	if in.byName[section] == nil {
		in.createValue(section, key, values)
		return
	}
	if in.byName[section].byKey[key] == nil {
		in.appendValue(section, key, values)
		return
	}
	in.replaceValue(section, key, values)
}

// createValue adds a new section and its first key.
func (in *Ini) createValue(section, key string, values []IniValue) {
	if len(values) == 0 {
		return
	}
	sec := &iniSection{index: len(in.sections), name: section, byKey: map[string]*iniEntry{}}
	in.sections = append(in.sections, sec)
	in.byName[section] = sec
	entry := &iniEntry{index: 0, key: key, values: values}
	sec.entries = append(sec.entries, entry)
	sec.byKey[key] = entry
	in.changed = true
}

// appendValue adds a new key to an existing section.
func (in *Ini) appendValue(section, key string, values []IniValue) {
	if len(values) == 0 {
		return
	}
	sec := in.byName[section]
	entry := &iniEntry{index: len(sec.entries), key: key, values: values}
	sec.entries = append(sec.entries, entry)
	sec.byKey[key] = entry
	in.changed = true
}

// replaceValue overwrites or deletes an existing key; deleting the last key
// removes the section too (upstream keeps empty sections out of serialize by
// the same values empty check).
func (in *Ini) replaceValue(section, key string, values []IniValue) {
	sec := in.byName[section]
	entry := sec.byKey[key]
	if len(values) == 0 {
		in.removeEntry(section, entry)
		return
	}
	if len(entry.values) == len(values) && equalValues(entry.values, values) {
		return
	}
	for i := range values {
		if i < len(entry.values) && values[i].Comment == nil && entry.values[i].Comment != nil {
			values[i].Comment = entry.values[i].Comment
		}
	}
	entry.values = values
	in.changed = true
}

// removeEntry drops one key and prunes a section left empty.
func (in *Ini) removeEntry(section string, entry *iniEntry) {
	sec := in.byName[section]
	delete(sec.byKey, entry.key)
	sec.entries = slices.DeleteFunc(sec.entries, func(e *iniEntry) bool { return e == entry })
	if len(sec.entries) == 0 {
		delete(in.byName, section)
		in.sections = slices.DeleteFunc(in.sections, func(s *iniSection) bool { return s == sec })
	}
	in.changed = true
}
