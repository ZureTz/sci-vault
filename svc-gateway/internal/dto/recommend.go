package dto

type RecommendSimilarQuery struct {
	LabID uint `form:"lab_id"`
	Limit int  `form:"limit" binding:"omitempty,min=1,max=20"`
}

type SimilarDocumentItem struct {
	DocID            uint     `json:"doc_id"`
	Title            string   `json:"title"`
	OriginalFileName string   `json:"original_file_name"`
	Summary          string   `json:"summary"`
	Authors          []string `json:"authors"`
	Tags             []string `json:"tags"`
	Similarity       float64  `json:"similarity"`
}

type RecommendSimilarResponse struct {
	Results []SimilarDocumentItem `json:"results"`
}

type RecommendForUserQuery struct {
	LabID uint `form:"lab_id"`
	Limit int  `form:"limit" binding:"omitempty,min=1,max=50"`
}

type RecommendForUserResponse struct {
	Results []SimilarDocumentItem `json:"results"`
}
