package namenode

import "fmt"

func createMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=create:%s@%s", object, node)
}

func commitMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=commit:%s@%s", object, node)
}

func deleteMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=delete:%s@%s", object, node)
}

func updateMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=update:%s@%s", object, node)
}
