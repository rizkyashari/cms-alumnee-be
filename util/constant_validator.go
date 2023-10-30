package util

func IsValidConstant[K comparable](constant K, constantMap map[K]bool) bool {
	_, ok := constantMap[constant]
	return ok
}
