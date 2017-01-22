package vfiles

type VFilesBox interface {
	FileSave(nameLocal, nameBox string) error
	FileGet(nameLocal, nameBox string) error
	FileRemove(nameBox string) error
}

// func FileSave(nameLocal, nameBox string) error {
// }
// func FileGet(nameLocal, nameBox string) error {
// }
// func FileRemove(nameBox string) error {
// }
