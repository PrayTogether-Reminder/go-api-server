package prayer

type PrayerService struct {
	repo *PrayerRepository
}

func NewPrayerService(repo *PrayerRepository) *PrayerService {
	return &PrayerService{repo: repo}
}
