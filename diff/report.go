package diff

// Report generates a flat list of differences encountered in the diff tree.
// Its output is less verbose than StringIndent as it doesn't report on
// matching values.
func Report(d Differ, outConf Output) ([]string, error) {
	var errs []string

	_, err := Walk(d, func(parent, diff Differ, path string) (Differ, error) {
		switch diff.Diff() {
		case Identical:
			return nil, nil
		case TypesDiffer:
			errs = append(errs, stringIndent(diff, " "+path+": ", "", outConf))
		case ContentDiffer:
			if _, ok := diff.(Walker); ok {
				return nil, nil
			}
			errs = append(errs, stringIndent(diff, " "+path+": ", "", outConf))
		}

		return nil, nil
	})

	return errs, err
}
