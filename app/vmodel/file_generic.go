package vmodel

// GetFidsAllowedOp returns a slice with the file IDs of the allowed operation
func GetEditedFidsAllowedOp(op string, reqUserID uint32, reqRole string, efids []uint32) []uint32 {
	rids := make([]uint32, 0, len(efids))
	for _, efid := range efids {
		if fdf, _, fus, err := GetFileEventUserFromEditedFileID(efid); err == nil {
			var allow bool
			if op == "delete" {
				allow = AllowDeleteEditedFile(reqUserID, fus, fdf)
			} else if op == "download" {
				allow = AllowDownloadEditedFile(reqUserID, fus, fdf)
			}
			if allow {
				rids = append(rids, efid)
			}
		}
	}
	return rids
}

func GetOriginalFidsAllowedOp(op string, reqUserID uint32, reqRole string, efids []uint32) []uint32 {
	rids := make([]uint32, 0, len(efids))
	for _, efid := range efids {
		if _, fev, fus, err := GetFileEventUserFromOriginalFileID(efid); err == nil {
			var allow bool
			if op == "delete" {
				allow = AllowDeleteOriginalFile(reqUserID, fus)
			} else if op == "download" {
				allow = AllowDownloadOriginalFile(reqUserID, reqRole, fus, fev)
			}
			if allow {
				rids = append(rids, efid)
			}
		}
	}
	return rids
}

func GetFileEventUserFromOriginalFileID(id uint32) (ef EventFile, ev Event, us User, err error) {
	ef, err = EventFileGetByEventFileID(id)
	if err != nil {
		return
	}
	ev, err = EventGetByEventID(ef.EventID)
	if err != nil {
		return
	}

	us, err = UserGetByID(ev.UserID)
	return
}

func GetFileEventUserFromEditedFileID(id uint32) (df EditedFile, ev Event, us User, err error) {
	df, err = EditedFileGetByEditedFileID(id)
	if err != nil {
		return
	}
	ev, err = EventGetByEventID(df.EventID)
	if err != nil {
		return
	}
	us, err = UserGetByID(ev.UserID)
	return
}

func AllowDeleteOriginalFile(reqUserID uint32, fileUser User) bool {
	if reqUserID == fileUser.ID {
		return true
	}
	return false
}

func AllowDeleteEditedFile(reqUserID uint32, fileUser User, df EditedFile) bool {
	if df.Status == "accepted" {
		// only the client-user can delete an accepted file
		if reqUserID == fileUser.ID {
			return true
		}
		return false
	}
	// only the editor can delete a non-accepted file
	if reqUserID == df.EditorID {
		return true
	}
	return false
}

// func AllowDeleteOriginalFileByID(reqUserID uint32, id uint32) (bool, error) {
// 	var err error
// 	if _, _, fus, err = GetFileEventUserFromOriginalFileID(id); err == nil {
// 		if allowDelete := AllowDeleteOriginalFile(reqUserID, fus); allowDelete != true {
// 			return true, nil
// 		}
// 	}
// 	return false, err
// }
// func AllowDeleteEditedFileByID(reqUserID uint32, id uint32) (bool, error) {
// 	var err error
// 	if fdf, _, fus, err = GetFileEventUserFromEditedFileID(id); err == nil {
// 		if allowDelete := AllowDeleteOriginalFile(reqUserID, fus, fdf); allowDelete != true {
// 			return true, nil
// 		}
// 	}
// 	return false, err
// }

// AllowDownloadOriginalFile
func AllowDownloadOriginalFile(reqUserID uint32, reqRole string, fileUser User, ev Event) bool {
	// it is assumes that reqUserID has already benn confirmed to have access to the event that the file belongs to
	return true
	// if reqUserID == fileUser.ID {
	// 	return true
	// }
	// // check if vsess.UserID is editor and if there is an association in editor_event
	// if reqRole != "editor" {
	// 	return false
	// }
	// _, err := EditorEventGetByEditorEventID(reqUserID, ev.ID)
	// if err == nil {
	// 	return true
	// }
	// return false
}

func AllowDownloadEditedFile(reqUserID uint32, fileUser User, df EditedFile) bool {
	if df.Status == "accepted" {
		// only the client-user can download an accepted file
		if reqUserID == fileUser.ID {
			return true
		}
		return false
	}
	// both client-user and the editor can delete a non-accepted file
	if reqUserID == fileUser.ID {
		return true
	}
	if reqUserID == df.EditorID {
		return true
	}
	return false
}
