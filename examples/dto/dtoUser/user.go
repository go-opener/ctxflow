package dtoUser

type AddUserReq struct {
    Name string  `json:"name" validate:"gt=0"`
    Age  *int32 `json:"age" validate:"omitempty,gt=0"`
}
