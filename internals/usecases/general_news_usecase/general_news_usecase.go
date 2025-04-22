package general_news_usecase

import (
	"VoteGolang/internals/data"
	"VoteGolang/internals/repositories/general_news_repository"
)

// GeneralNewsUseCase handles logic for retrieving general election news.
type GeneralNewsUseCase struct {
	Repo general_news_repository.GeneralNewsRepository
}

func NewGeneralNewsUseCase(repo general_news_repository.GeneralNewsRepository) *GeneralNewsUseCase {
	return &GeneralNewsUseCase{Repo: repo}
}

func (uc *GeneralNewsUseCase) Create(news *data.GeneralNews) error {
	return uc.Repo.Create(news)
}

// GetAll returns a list of general news articles.
func (uc *GeneralNewsUseCase) GetAll() ([]data.GeneralNews, error) {
	return uc.Repo.GetAll()
}

func (uc *GeneralNewsUseCase) GetByID(id uint) (*data.GeneralNews, error) {
	return uc.Repo.GetByID(id)
}

func (uc *GeneralNewsUseCase) Update(news *data.GeneralNews) error {
	return uc.Repo.Update(news)
}

func (uc *GeneralNewsUseCase) Delete(id uint) error {
	return uc.Repo.Delete(id)
}
