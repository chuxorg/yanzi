package cmd

import "strings"

const protocolAnnotationSemantics = "annotation_only"

type protocolAnnotation struct {
	Raw        string
	Kind       string
	Argument   string
	Executable bool
	Semantics  string
}

func parseProtocolAnnotation(raw string) (protocolAnnotation, bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return protocolAnnotation{}, false
	}

	lowered := strings.ToLower(trimmed)
	if !strings.HasPrefix(lowered, "@yanzi") {
		return protocolAnnotation{}, false
	}

	remainder := strings.TrimSpace(trimmed[len("@yanzi"):])
	if remainder == "" {
		return protocolAnnotation{}, false
	}

	kind := remainder
	argument := ""
	if idx := strings.IndexAny(remainder, " \t"); idx >= 0 {
		kind = remainder[:idx]
		argument = strings.TrimSpace(remainder[idx+1:])
	}
	kind = strings.ToLower(strings.TrimSpace(kind))
	switch kind {
	case "pause", "resume", "checkpoint", "export", "role":
	default:
		kind = "custom"
	}

	return protocolAnnotation{
		Raw:        trimmed,
		Kind:       kind,
		Argument:   trimProtocolArgument(argument),
		Executable: false,
		Semantics:  protocolAnnotationSemantics,
	}, true
}

func trimProtocolArgument(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"")) || (strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'")) {
			return strings.TrimSpace(value[1 : len(value)-1])
		}
	}
	return value
}

func protocolKindLabel(kind string) string {
	switch kind {
	case "pause":
		return "pause"
	case "resume":
		return "resume"
	case "checkpoint":
		return "checkpoint"
	case "export":
		return "export"
	case "role":
		return "role"
	default:
		return "custom"
	}
}
