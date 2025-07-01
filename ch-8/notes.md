Copying over the explanation for resilience:
// Resilience note:
// Commit the offsets after stats are updated and saved via UpdateStats batching.
// If the app crashes or is restarted before the commit, Kafka should process the messages.