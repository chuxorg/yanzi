// Package hash provides deterministic intent hashing logic.
package hash

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/chuxorg/yanzi/internal/core/model"
)

// CanonicalizeMeta re-encodes a JSON object with sorted keys.
func CanonicalizeMeta(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return nil, errors.New("meta must be a JSON object")
	}

	var b strings.Builder
	if err := writeJSONObject(&b, obj); err != nil {
		return nil, err
	}
	return json.RawMessage(b.String()), nil
}

// HashIntent computes a deterministic SHA-256 hash for an IntentRecord.
// The hash preimage excludes the hash field and uses canonical field order.
func HashIntent(record model.IntentRecord) (string, error) {
	normalized := record.Normalize()
	preimage, err := canonicalIntentPreimage(normalized)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(preimage)
	return hex.EncodeToString(sum[:]), nil
}

func canonicalIntentPreimage(record model.IntentRecord) ([]byte, error) {
	if len(record.ID) == 0 {
		return nil, errors.New("id is required for hashing")
	}
	if len(record.CreatedAt) == 0 {
		return nil, errors.New("created_at is required for hashing")
	}
	createdAt, err := normalizeRFC3339(record.CreatedAt)
	if err != nil {
		return nil, errors.New("created_at must be RFC3339")
	}
	if len(record.Author) == 0 {
		return nil, errors.New("author is required for hashing")
	}
	if len(record.SourceType) == 0 {
		return nil, errors.New("source_type is required for hashing")
	}
	if len(record.Prompt) == 0 {
		return nil, errors.New("prompt is required for hashing")
	}
	if len(record.Response) == 0 {
		return nil, errors.New("response is required for hashing")
	}

	var meta json.RawMessage
	if len(record.Meta) > 0 {
		canonicalMeta, err := CanonicalizeMeta(record.Meta)
		if err != nil {
			return nil, err
		}
		meta = canonicalMeta
	}

	var b strings.Builder
	b.WriteByte('{')
	first := true

	addStringField(&b, &first, "id", record.ID)
	addStringField(&b, &first, "created_at", createdAt)
	addStringField(&b, &first, "author", record.Author)
	addStringField(&b, &first, "source_type", record.SourceType)
	if record.Title != "" {
		addStringField(&b, &first, "title", record.Title)
	}
	addStringField(&b, &first, "prompt", record.Prompt)
	addStringField(&b, &first, "response", record.Response)
	if len(meta) > 0 {
		addRawField(&b, &first, "meta", meta)
	}
	if record.PrevHash != "" {
		addStringField(&b, &first, "prev_hash", record.PrevHash)
	}
	b.WriteByte('}')

	return []byte(b.String()), nil
}

func normalizeRFC3339(value string) (string, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", err
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

func addStringField(b *strings.Builder, first *bool, name string, value string) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	encoded, _ := json.Marshal(value)
	b.Write(encoded)
}

func addRawField(b *strings.Builder, first *bool, name string, raw json.RawMessage) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	b.Write(raw)
}

func decodeJSON(raw json.RawMessage) (any, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if err := ensureEOF(dec); err != nil {
		return nil, err
	}
	return v, nil
}

func ensureEOF(dec *json.Decoder) error {
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return errors.New("unexpected trailing JSON data")
	} else if !errors.Is(err, io.EOF) {
		return errors.New("unexpected trailing JSON data")
	}
	return nil
}

func writeJSONObject(b *strings.Builder, obj map[string]any) error {
	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	b.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		encodedKey, _ := json.Marshal(key)
		b.Write(encodedKey)
		b.WriteByte(':')
		if err := writeJSONValue(b, obj[key]); err != nil {
			return err
		}
	}
	b.WriteByte('}')
	return nil
}

func writeJSONValue(b *strings.Builder, value any) error {
	switch v := value.(type) {
	case nil:
		b.WriteString("null")
	case bool:
		if v {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case string:
		encoded, _ := json.Marshal(v)
		b.Write(encoded)
	case json.Number:
		b.WriteString(v.String())
	case float64:
		b.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
	case []any:
		b.WriteByte('[')
		for i, item := range v {
			if i > 0 {
				b.WriteByte(',')
			}
			if err := writeJSONValue(b, item); err != nil {
				return err
			}
		}
		b.WriteByte(']')
	case map[string]any:
		if err := writeJSONObject(b, v); err != nil {
			return err
		}
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return err
		}
		b.Write(encoded)
	}
	return nil
}
