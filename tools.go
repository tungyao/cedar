package ultimate_cedar

func inArrayString(target string, srcArr []string) bool {
	for _, v := range srcArr {
		if v == target {
			return true
		}
	}
	return false
}
