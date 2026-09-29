package agent

func numberAsInt64(
	value any,
) (int64, bool) {
	switch v := value.(type) {
	case float64:
		return int64(v), true

	case float32:
		return int64(v), true

	case int:
		return int64(v), true

	case int8:
		return int64(v), true

	case int16:
		return int64(v), true

	case int32:
		return int64(v), true

	case int64:
		return v, true

	case uint:
		return int64(v), true

	case uint8:
		return int64(v), true

	case uint16:
		return int64(v), true

	case uint32:
		return int64(v), true

	case uint64:
		return int64(v), true

	default:
		return 0, false
	}
}
