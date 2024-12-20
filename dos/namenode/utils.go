package namenode

import "fmt"

func CreateMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=create:%s@%s", object, node)
}

func CommitMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=commit:%s@%s", object, node)
}

func DeleteMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=delete:%s@%s", object, node)
}

func UpdateMessageTag(object string, node string) string {
	return fmt.Sprintf("$tag=update:%s@%s", object, node)
}

func DistributedReadTag(objects []string, node string) string {
	return fmt.Sprintf("$tag=distributed-read:%d-objects@%s", len(objects), node)
}
