package integration

import (
	"bytes"
	"fmt"

	"github.com/tailscale/hujson"
)

// JSON config editing built on hujson (JWCC), so comments, trailing commas and
// formatting in a user's settings file survive an edit. Every JSON-editing
// integration goes through here.

// setJSONPath sets a string at a nested path, creating missing objects on the
// way down.
func setJSONPath(config string, path []string, value string) (string, error) {
	return editJSON(config, func(root *hujson.Value) error {
		parent, err := objectAt(root, path[:len(path)-1], true)
		if err != nil {
			return err
		}
		setMember(parent, path[len(path)-1], hujson.Value{Value: hujson.String(value)})
		return nil
	})
}

// deleteJSONPath removes a key. A path that does not exist is a no-op, so reset
// stays idempotent.
func deleteJSONPath(config string, path []string) (string, error) {
	return editJSON(config, func(root *hujson.Value) error {
		parent, err := objectAt(root, path[:len(path)-1], false)
		if err != nil || parent == nil {
			return err
		}
		i := memberIndex(parent, path[len(path)-1])
		if i < 0 {
			return nil
		}
		trailing := parent.Members[i].Value.AfterExtra
		parent.Members = append(parent.Members[:i], parent.Members[i+1:]...)
		// removing the last member moves the trailing comma to its predecessor
		if n := len(parent.Members); i == n && n > 0 {
			parent.Members[n-1].Value.AfterExtra = trailing
		}
		return nil
	})
}

// upsertJSONArrayByName replaces the array element whose "name" matches, or
// appends obj when there is none. Elements themectl does not own are untouched.
func upsertJSONArrayByName(config string, path []string, name string, obj []byte) (string, error) {
	parsed, err := hujson.Parse(obj)
	if err != nil {
		return "", fmt.Errorf("parse object for %q: %w", name, err)
	}

	return editJSON(config, func(root *hujson.Value) error {
		arr, err := arrayAt(root, path, true)
		if err != nil {
			return err
		}
		if i := elementIndex(arr, name); i >= 0 {
			// keep the element's surrounding whitespace and comments
			parsed.BeforeExtra = arr.Elements[i].BeforeExtra
			parsed.AfterExtra = arr.Elements[i].AfterExtra
			arr.Elements[i] = parsed
			return nil
		}
		parsed.BeforeExtra = hujson.Extra(elementIndent(arr))
		// hujson records a trailing comma on the last element only, so carry it
		// onto the new one or an append silently drops it
		if n := len(arr.Elements); n > 0 {
			parsed.AfterExtra = arr.Elements[n-1].AfterExtra
		}
		arr.Elements = append(arr.Elements, parsed)
		return nil
	})
}

// removeJSONArrayByName drops the element whose "name" matches. Missing array
// or missing element is a no-op.
func removeJSONArrayByName(config string, path []string, name string) (string, error) {
	return editJSON(config, func(root *hujson.Value) error {
		arr, err := arrayAt(root, path, false)
		if err != nil || arr == nil {
			return err
		}
		i := elementIndex(arr, name)
		if i < 0 {
			return nil
		}
		trailing := arr.Elements[i].AfterExtra
		arr.Elements = append(arr.Elements[:i], arr.Elements[i+1:]...)
		// removing the last element moves the trailing comma to its predecessor
		if n := len(arr.Elements); i == n && n > 0 {
			arr.Elements[n-1].AfterExtra = trailing
		}
		return nil
	})
}

func editJSON(config string, edit func(*hujson.Value) error) (string, error) {
	root, err := hujson.Parse([]byte(config))
	if err != nil {
		return "", fmt.Errorf("parse json config: %w", err)
	}
	if _, ok := root.Value.(*hujson.Object); !ok {
		return "", fmt.Errorf("no object found in config")
	}
	if err := edit(&root); err != nil {
		return "", err
	}
	return string(root.Pack()), nil
}

// objectAt walks path from root. With create, missing or non-object links are
// replaced by empty objects; without it, a missing link returns a nil object.
func objectAt(v *hujson.Value, path []string, create bool) (*hujson.Object, error) {
	obj, ok := v.Value.(*hujson.Object)
	if !ok {
		return nil, fmt.Errorf("no object found in config")
	}

	for _, key := range path {
		idx := memberIndex(obj, key)
		if idx < 0 {
			if !create {
				return nil, nil
			}
			setMember(obj, key, hujson.Value{Value: &hujson.Object{}})
			idx = memberIndex(obj, key)
		}

		next, ok := obj.Members[idx].Value.Value.(*hujson.Object)
		if !ok {
			if !create {
				return nil, nil
			}
			// a non-object here cannot be descended into; replacing it is the
			// only way to set the requested path
			return nil, fmt.Errorf("%q is not an object", key)
		}
		obj = next
	}
	return obj, nil
}

func arrayAt(v *hujson.Value, path []string, create bool) (*hujson.Array, error) {
	parent, err := objectAt(v, path[:len(path)-1], create)
	if err != nil || parent == nil {
		return nil, err
	}

	key := path[len(path)-1]
	idx := memberIndex(parent, key)
	if idx < 0 {
		if !create {
			return nil, nil
		}
		setMember(parent, key, hujson.Value{Value: &hujson.Array{}})
		idx = memberIndex(parent, key)
	}

	arr, ok := parent.Members[idx].Value.Value.(*hujson.Array)
	if !ok {
		if !create {
			return nil, nil
		}
		return nil, fmt.Errorf("%q is not an array", key)
	}
	return arr, nil
}

func memberIndex(obj *hujson.Object, key string) int {
	want := string(hujson.String(key))
	for i, m := range obj.Members {
		if lit, ok := m.Name.Value.(hujson.Literal); ok && string(lit) == want {
			return i
		}
	}
	return -1
}

// setMember replaces a member's value in place, keeping its position and
// surrounding comments, or appends it indented like its new siblings.
func setMember(obj *hujson.Object, key string, val hujson.Value) {
	if i := memberIndex(obj, key); i >= 0 {
		val.BeforeExtra = obj.Members[i].Value.BeforeExtra
		val.AfterExtra = obj.Members[i].Value.AfterExtra
		obj.Members[i].Value = val
		return
	}

	indent := memberIndent(obj)
	empty := len(obj.Members) == 0
	val.BeforeExtra = hujson.Extra(" ")
	// hujson records a trailing comma on the last member only, so carry it onto
	// the new one or an append silently drops it
	if !empty {
		val.AfterExtra = obj.Members[len(obj.Members)-1].Value.AfterExtra
	}
	obj.Members = append(obj.Members, hujson.ObjectMember{
		Name:  hujson.Value{BeforeExtra: hujson.Extra(indent), Value: hujson.String(key)},
		Value: val,
	})
	// an object that already had members has its closing brace placed the way
	// the user wrote it; only give a fresh one somewhere sensible to sit
	if empty {
		obj.AfterExtra = hujson.Extra(closingIndent(indent))
	}
}

func elementIndex(arr *hujson.Array, name string) int {
	want := string(hujson.String(name))
	for i, el := range arr.Elements {
		obj, ok := el.Value.(*hujson.Object)
		if !ok {
			continue
		}
		j := memberIndex(obj, "name")
		if j < 0 {
			continue
		}
		if lit, ok := obj.Members[j].Value.Value.(hujson.Literal); ok && string(lit) == want {
			return i
		}
	}
	return -1
}

// memberIndent reuses the whitespace of an existing member so an appended key
// lines up with its siblings.
func memberIndent(obj *hujson.Object) string {
	for _, m := range obj.Members {
		if in := lastLineIndent(m.Name.BeforeExtra); in != "" {
			return in
		}
	}
	return "\n  "
}

// elementIndent matches an existing element's indentation. An array written on
// one line stays on one line.
func elementIndent(arr *hujson.Array) string {
	for _, el := range arr.Elements {
		if in := lastLineIndent(el.BeforeExtra); in != "" {
			return in
		}
	}
	return ""
}

// lastLineIndent returns the newline and indent that directly precede a value,
// ignoring any comments earlier in the extra.
func lastLineIndent(extra hujson.Extra) string {
	i := bytes.LastIndexByte(extra, '\n')
	if i < 0 {
		return ""
	}
	rest := extra[i+1:]
	if len(bytes.TrimLeft(rest, " \t")) != 0 {
		return "" // a comment follows the newline, not plain indent
	}
	return string(extra[i:])
}

func closingIndent(memberIndent string) string {
	i := bytes.LastIndexByte([]byte(memberIndent), '\n')
	if i < 0 {
		return "\n"
	}
	body := memberIndent[i+1:]
	if len(body) >= 2 {
		return memberIndent[:i+1] + body[2:]
	}
	return memberIndent[:i+1]
}
