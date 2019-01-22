package passportChecker

type MultiChecker struct {
	soft   ExistChecker
	strong ExistChecker
}

func MakeMultiChecker(soft, strong ExistChecker) *MultiChecker {
	return &MultiChecker{soft, strong}
}

func (mc *MultiChecker) Add(values []interface{}) error {
	existsMap, err := mc.Check(values)
	if err != nil {
		return err
	}
	toAdd := make([]interface{}, 0)
	for i, e := range existsMap {
		if !e {
			toAdd = append(toAdd, values[i])
		}
	}
	if len(toAdd) == 0 {
		return nil
	}
	errCh := make(chan error)
	defer close(errCh)
	go func() {
		errCh <- mc.soft.Add(toAdd)
	}()
	go func() {
		errCh <- mc.strong.Add(toAdd)
	}()

	if err := <-errCh; err != nil {
		go func() {
			<-errCh
			close(errCh)
		}()
		return err
	}
	return <-errCh

}

func (mc *MultiChecker) Check(values []interface{}) ([]bool, error) {
	if len(values) == 0 {
		return make([]bool, 0), nil
	}
	existsMap, err := mc.soft.Check(values)
	if err != nil {
		return []bool{}, err
	}
	valuesInSoft := make([]interface{}, 0)
	valuesInSoftMap := make(map[int]int, 0)
	j := 0
	for i, e := range existsMap {
		if e {
			valuesInSoft = append(valuesInSoft, values[i])
			valuesInSoftMap[j] = i
		}
		j++
	}
	if len(valuesInSoft) == 0 {
		return existsMap, nil
	}
	existsStrong, err := mc.strong.Check(valuesInSoft)
	if err != nil {
		return []bool{}, err
	}
	for j, e := range existsStrong {
		existsMap[valuesInSoftMap[j]] = e
	}
	return existsMap, nil
}
