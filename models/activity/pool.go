package activity

type Pool struct {
}

func (p *Pool) NewActivity() *Activity {
	activity := Activity{}
	return &activity
}
