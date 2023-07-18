package errors

type metadata struct {
	key   string
	value interface{}
}

type MetaProvider interface {
	Meta(key string) interface{}
	MetaMap() map[string]interface{}
}

func Meta(key string, value interface{}) interface{} {
	return metadata{key: key, value: value}
}

func extractMetaFrom(args ...interface{}) (map[string]interface{}, []interface{}) {
	meta := make(map[string]interface{})
	var newArgs []interface{}
	for _, arg := range args {
		if m, ok := arg.(metadata); ok {
			meta[m.key] = m.value
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	return meta, newArgs
}

func GetMeta(err error, key string) interface{} {
	for err != nil {
		if mp, ok := err.(MetaProvider); ok {
			if v, ok := mp.MetaMap()[key]; ok {
				return v
			}
		}
		err = Unwrap(err)
	}
	return nil
}
