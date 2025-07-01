package dao

import "strings"

func convertBool(val bool) string {
	res := "N"

	if val {
		res = "Y"
	}

	return res
}

func convertString(val string) bool {
	uval := strings.ToUpper(val)
	return uval == "Y"
}

func prepSearch(val string) string {
	res := val
	if !strings.HasPrefix(res, "%") {
		res = "%" + res
	}

	if !strings.HasSuffix(res, "%") {
		res = res + "%"
	}

	// now internal, ? and * should be processed as well
	res = strings.ReplaceAll(res, "?", "_")
	res = strings.ReplaceAll(res, "*", "%")

	return res
}

func getStringParams(ids []string) []any {
	var res []any
	for _, id := range ids {
		res = append(res, id)
	}

	return res
}

func getFirst(first bool, parts []string, toAdd string) (bool, []string) {
	if !first {
		parts = append(parts, toAdd)
	}

	return false, parts
}
