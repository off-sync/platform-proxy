package startwatcher

// Command defines the StartWatcher command.
// This command starts a watcher for config changes. The watcher should stop
// when the Done channel of the command model context object is closed.
// The callback function should be called when a change in configuration is observed.
type Command interface {
	Execute(model *CommandModel) error
}
