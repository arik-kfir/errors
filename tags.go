package errors

type tag struct {
	tag string
}

type TagProvider interface {
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
		if mp, ok := err.(TagProvider); ok {
			for _, t := range mp.Tags() {
				if t == tag {
					return true
				}
			}
		}
		err = Unwrap(err)
	}
	return false
}
