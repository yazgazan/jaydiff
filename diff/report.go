package diff

type ReportError string

func (e ReportError) Error() string {
	return string(e)
}

func Report(d Differ, outConf Output) ([]error, error) {
	var errs []error

	err := Walk(d, func(parent, diff Differ, path string) error {
		switch diff.Diff() {
		case Identical:
			return nil
		case TypesDiffer:
			errs = append(errs, ReportError(diff.StringIndent(" "+path+": ", "", outConf)))
		case ContentDiffer:
			if _, ok := diff.(Walker); ok {
				return nil
			}
			errs = append(errs, ReportError(diff.StringIndent(" "+path+": ", "", outConf)))
		}

		return nil
	})

	return errs, err
}
