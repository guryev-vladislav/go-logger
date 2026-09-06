package handlers

import (
	"fmt"
	"strings"
	"time"
)

func FormatSIPBlock(direction string, rawMessage string, timestamp time.Time) string {
	headerColor, headerSymbol, headerText := getHeaderParams(direction)

	timestampStr := timestamp.Format(timeFormat)
	header := fmt.Sprintf(headerPrefixColored,
		headerColor, headerSymbol, headerText, timestampStr, colorReset)

	body := HighlightSIPMessage(rawMessage)
	footer := fmt.Sprintf("%s%s%s\n", colorGray, separatorLine, colorReset)

	return header + newline + body + newline + footer
}

func FormatSIPBlockPlain(direction string, rawMessage string, timestamp time.Time) string {
	headerText := getPlainHeaderText(direction)
	timestampStr := timestamp.Format(timeFormat)

	return fmt.Sprintf(headerPrefixPlain,
		headerText, timestampStr, rawMessage, separatorLine)
}

func getHeaderParams(direction string) (string, string, string) {
	switch strings.ToUpper(direction) {
	case directionSent:
		return colorCyan, symbolSent, headerTextSent
	default:
		return colorMagenta, symbolReceived, headerTextReceived
	}
}

func getPlainHeaderText(direction string) string {
	switch strings.ToUpper(direction) {
	case directionSent:
		return symbolSent + space + headerTextSent
	default:
		return symbolReceived + space + headerTextReceived
	}
}

func HighlightSIPMessage(rawMessage string) string {
	if rawMessage == emptyString {
		return emptyString
	}

	lines := strings.Split(strings.TrimRight(rawMessage, newline), newline)

	var builder strings.Builder
	builder.Grow(len(rawMessage))

	for i, line := range lines {
		if i > 0 {
			if _, err := builder.WriteString(newline); err != nil {
				return builder.String()
			}
		}

		if _, err := builder.WriteString(HighlightSIPLine(line)); err != nil {
			return builder.String()
		}
	}

	return builder.String()
}

func HighlightSIPLine(line string) string {
	trimmedLine := strings.TrimLeft(line, space)
	for _, method := range sipMethods {
		if strings.HasPrefix(trimmedLine, method+space) {
			methodPos := strings.Index(line, method)
			if methodPos >= 0 {
				return line[:methodPos] + fmt.Sprintf("%s%s%s", colorCyan, method, colorReset) + line[methodPos+len(method):]
			}
		}
	}

	if strings.HasPrefix(line, sipVersion) {
		return highlightSIPStatusLine(line)
	}

	lineLower := strings.ToLower(line)
	for _, header := range sipHeaders {
		headerLower := strings.ToLower(header)
		if strings.HasPrefix(lineLower, headerLower) {
			return fmt.Sprintf("%s%s%s%s", colorYellow, line[:len(header)], colorReset, line[len(header):])
		}
	}

	return line
}

func highlightSIPStatusLine(line string) string {
	parts := strings.Fields(line)

	for i, part := range parts {
		if isStatusCode(part) {
			parts[i] = ColorizeStatusCode(part)
		}
	}

	return strings.Join(parts, space)
}

func isStatusCode(status string) bool {
	if len(status) != 3 {
		return false
	}

	for _, ch := range status {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return status[0] >= '1' && status[0] <= '6'
}

func ColorizeStatusCode(code string) string {
	if !isStatusCode(code) {
		return code
	}

	switch code[0] {
	case '1':
		return fmt.Sprintf("%s%s%s", colorGray, code, colorReset)
	case '2':
		return fmt.Sprintf("%s%s%s", colorGreen, code, colorReset)
	case '3':
		return fmt.Sprintf("%s%s%s", colorBlue, code, colorReset)
	case '4':
		return fmt.Sprintf("%s%s%s", colorYellow, code, colorReset)
	case '5', '6':
		return fmt.Sprintf("%s%s%s", colorRed, code, colorReset)
	default:
		return code
	}
}

func ExtractCallIDFromSIP(rawMessage string) string {
	if rawMessage == emptyString {
		return emptyString
	}

	lines := strings.SplitSeq(rawMessage, newline)
	for line := range lines {
		lineLower := strings.ToLower(line)
		callIDLower := strings.ToLower(callIDPrefix)

		if strings.HasPrefix(lineLower, callIDLower) {
			value := strings.TrimPrefix(line, callIDPrefix)
			if after, ok := strings.CutPrefix(line, callIDPrefix+" "); ok {
				value = after
			}

			return strings.TrimSpace(value)
		}
	}

	return emptyString
}

func ExtractMethodFromSIP(rawMessage string) string {
	if rawMessage == emptyString {
		return emptyString
	}

	firstLine := getFirstLine(rawMessage)
	if firstLine == emptyString {
		return emptyString
	}

	if strings.HasPrefix(firstLine, sipVersion) {
		return emptyString
	}

	parts := strings.Fields(firstLine)
	if len(parts) > 0 {
		return parts[0]
	}

	return emptyString
}

type SIPFirstLine struct {
	Method     string
	RequestURI string
	StatusCode int
	StatusText string
	Version    string
}

func ParseFirstLine(line string) SIPFirstLine {
	result := SIPFirstLine{}

	if line == emptyString {
		return result
	}

	parts := strings.Fields(line)
	if len(parts) == 0 {
		return result
	}

	if strings.HasPrefix(line, sipVersion) {
		result.Version = parts[0]
		if len(parts) >= 2 {
			var statusCode int
			if _, err := fmt.Sscanf(parts[1], "%d", &statusCode); err == nil {
				result.StatusCode = statusCode
			}

			if len(parts) >= 3 {
				result.StatusText = strings.Join(parts[2:], space)
			}
		}
	} else {
		if len(parts) >= 1 {
			result.Method = parts[0]
		}

		if len(parts) >= 2 {
			result.RequestURI = parts[1]
		}

		if len(parts) >= 3 {
			result.Version = parts[2]
		}
	}

	return result
}

func ExtractStatusFromSIP(rawMessage string) (int, string) {
	if rawMessage == emptyString {
		return 0, emptyString
	}

	firstLine := getFirstLine(rawMessage)
	if firstLine == emptyString {
		return 0, emptyString
	}

	if !strings.HasPrefix(firstLine, sipVersion) {
		return 0, emptyString
	}

	parsed := ParseFirstLine(firstLine)

	return parsed.StatusCode, parsed.StatusText
}

func getFirstLine(rawMessage string) string {
	if rawMessage == emptyString {
		return emptyString
	}

	lines := strings.Split(rawMessage, newline)
	if len(lines) == 0 {
		return emptyString
	}

	return lines[0]
}
