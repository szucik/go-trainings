package main

import (
	"regexp"
	"strconv"
	"strings"
)

package alarmcleaner

import (
    "regexp"
    "strconv"
    "strings"
)

// Końcowa funkcja – użyj tylko tej jednej!
func BeautifyAndNormalize(rule string) string {
    // 1. Najpierw znajdź każdy ALARM(...)
    re := regexp.MustCompile(`ALARM\s*\([^)]*\)`)

    result := re.ReplaceAllStringFunc(rule, func(part string) string {
        // Wydłub wnętrze
        inner := strings.TrimPrefix(part, "ALARM")
        inner = strings.Trim(inner, "()")
        inner = strings.TrimSpace(inner)

        // Spróbuj zinterpretować jako string Go (obsługuje \"nazwa\, "prefix"+var itd.)
        if unquoted, err := strconv.Unquote(`"` + inner + `"`); err == nil {
            name := strings.TrimSpace(unquoted)
            name = strings.Trim(name, `"`)
            return `ALARM('` + name + `')`
        }

        // Fallback – ręczne czyszczenie (na wszelki wypadek)
        name := inner
        name = strings.ReplaceAll(name, `\"`, `"`)
        name = strings.Trim(name, `"`)
        name = strings.TrimSpace(name)
        return `ALARM('` + name + `')`
    })

    // 2. Teraz czyścimy całą regułę – usuwamy zbędne białe znaki
    result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
    result = strings.TrimSpace(result)

    // 3. Opcjonalnie: dodaj spację po AND/OR/NOT dla piękna
    result = regexp.MustCompile(`\s+(AND|OR|NOT)\s+`).ReplaceAllStringFunc(result, func(s string) string {
        return " " + strings.TrimSpace(s) + " "
    })

    return strings.TrimSpace(result)
}
