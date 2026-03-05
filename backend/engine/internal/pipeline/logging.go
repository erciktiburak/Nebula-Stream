package pipeline

func LogEvent(topic string) string {
  return "ingestion:" + topic
}
