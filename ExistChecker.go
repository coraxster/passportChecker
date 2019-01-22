package passportChecker

//попробовать https://github.com/HouzuoGuo/tiedot
//вместо sqlite, возмодно побыстрее
//mariadb точно медленнее(локально на буке)

// на будущее, рассчитывать вероятность в %, задавать необзодимую в MultiStore
// принимать []Store, находить и суммировать вероятности, пока не найдём необходимую
// позволит комбинировать []Store

type ExistChecker interface {
	Add([]interface{}) error
	Check([]interface{}) ([]bool, error)
}
