package docs

//go:generate swag init -g apiserver.go -d ../internal/cmd,../internal/apiserver,../pkg/types -o ../docs
