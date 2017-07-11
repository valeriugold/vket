package vmodel

import "github.com/valeriugold/vket/app/shared/database"

func EventFileGetForUserIDEventID(userID, eventID uint32, role string) ([]EventFile, error) {
	var err error

	var result []EventFile
	switch database.ReadConfig().Type {
	case database.TypeMySQL:
		if role == "user" {
			err = database.SQL.Select(&result, "SELECT a.id id, a.event_id event_id, a.owner_id owner_id, "+
				"a.status status, a.name name, a.created_at created_at, a.updated_at updated_at FROM "+
				"event_file a "+
				"INNER JOIN event b on a.event_id=b.id "+
				"INNER JOIN user c on b.user_id=c.id "+
				"LEFT JOIN editor_event d on a.event_id=d.event_id "+
				"WHERE b.id = ? AND b.user_id = ? AND (a.status = 'original' OR a.status = 'preview')",
				eventID, userID)
		} else if role == "editor" {
			err = database.SQL.Select(&result, "SELECT a.id id, a.event_id event_id, a.owner_id owner_id, "+
				"a.status status, a.name name, a.created_at created_at, a.updated_at updated_at FROM "+
				"event_file a "+
				"INNER JOIN event b on a.event_id=b.id "+
				"INNER JOIN user c on b.user_id=c.id "+
				"LEFT JOIN editor_event d on a.event_id=d.event_id "+
				"WHERE b.id = ? AND d.editor_id = ?",
				eventID, userID)
		}
	default:
		err = ErrCode
	}

	return result, standardizeError(err)
}

func GetEventFileIDsAllowedDelete(reqUserID uint32, reqRole string, efids []uint32) []uint32 {
	rids := make([]uint32, 0, len(efids))
	for _, efid := range efids {
		// ef, ev, ou, evus, err := GetFileEventOwnerUserFromOriginalFileID(efid)
		_, _, ou, _, err := GetFileEventOwnerUserFromOriginalFileID(efid)
		if err == nil {
			if reqUserID == ou.ID {
				rids = append(rids, efid)
			}
		}
	}
	return rids
}

func GetEventFileIDsAllowedDownload(reqUserID uint32, reqRole string, efids []uint32) []uint32 {
	rids := make([]uint32, 0, len(efids))
	for _, efid := range efids {
		_, ev, ou, _, err := GetFileEventOwnerUserFromOriginalFileID(efid)
		if err == nil {
			if reqUserID == ou.ID {
				rids = append(rids, efid)
				continue
			}
			// for user check if it owns the event
			if reqRole == "user" {
				if reqUserID == ev.UserID {
					if ev.Status == "preview" {
						rids = append(rids, efid)
					}
				}
				continue
			}
			// for editor, check if it has access to event and file is "original"
			if reqRole == "editor" {
				if ee, err := EditorEventGetByEditorEventID(reqUserID, ev.ID); err != nil {
					if ee.ID != 0 {
						// there is an association between user and event
						rids = append(rids, efid)
						continue
					}
				}
			}
		}
	}
	return rids
}

func GetEventFileIDsAllowedAccept(reqUserID uint32, reqRole string, efids []uint32) []uint32 {
	rids := make([]uint32, 0, len(efids))
	if reqRole == "editor" {
		return rids
	}
	for _, efid := range efids {
		if ef, ev, _, _, err := GetFileEventOwnerUserFromOriginalFileID(efid); err == nil {
			if ef.Status == "preview" && reqUserID == ev.UserID {
				rids = append(rids, efid)
				// // get the actual proposal file, not the preview one
				// if af, err := EventFileGetProposal(ef); err == nil {
				// 	rids = append(rids, af.ID)
				// }
			}
		}
	}
	return rids
}

func GetEventFileIDsAllowedReject(reqUserID uint32, reqRole string, efids []uint32) []uint32 {
	return GetEventFileIDsAllowedAccept(reqUserID, reqRole, efids)
}

func GetFileEventOwnerUserFromOriginalFileID(id uint32) (ef EventFile, ev Event, ou User, us User, err error) {
	ef, err = EventFileGetByEventFileID(id)
	if err != nil {
		return
	}
	ev, err = EventGetByEventID(ef.EventID)
	if err != nil {
		return
	}
	if ef.OwnerID != 0 {
		ou, err = UserGetByID(ef.OwnerID)
		if err != nil {
			return
		}
	}
	us, err = UserGetByID(ev.UserID)
	return
}
