package utils

// 对userIDs进行去重
func RemoveDuplicate(userIDs []string) []string {
	userIDMap := make(map[string]struct{})
	for _, userID := range userIDs {
		userIDMap[userID] = struct{}{}
	}

	userIDs = make([]string, 0, len(userIDMap))
	for userID := range userIDMap {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}
