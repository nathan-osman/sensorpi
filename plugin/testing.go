package plugin

func IsInputPlugin(v any) bool {
	_, ok := v.(InputPlugin)
	return ok
}

func IsOutputPlugin(v any) bool {
	_, ok := v.(OutputPlugin)
	return ok
}

func IsTriggerPlugin(v any) bool {
	_, ok := v.(TriggerPlugin)
	return ok
}
