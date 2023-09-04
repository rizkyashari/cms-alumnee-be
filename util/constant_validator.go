package util

func IsValidConstant(constant int, constantMap map[int]bool) bool {
	_, ok := constantMap[constant]
	return ok
}
