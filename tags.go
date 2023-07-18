package errors

type tag struct {
	tag string
}

type TagProvider interface {
	HasTag(tag string) bool
	Tags() []string
}

func Tag(t string) interface{} {
	return tag{tag: t}
}

func extractTagsFrom(args ...interface{}) ([]string, []interface{}) {
	var tags []string
	var newArgs []interface{}
	for _, arg := range args {
		if t, ok := arg.(tag); ok {
			tags = append(tags, t.tag)
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	return tags, newArgs
}

func HasTag(err error, tag string) bool {
	for err != nil {
		if tp, ok := err.(TagProvider); ok {
			if tp.HasTag(tag) {
				return true
			}
		}
		err = Unwrap(err)
	}
	return false
}
