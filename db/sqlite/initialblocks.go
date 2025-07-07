package sqlite

import (
	"mapserver/db"
	"mapserver/settings"
	"mapserver/types"
)

const (
	SETTING_LAST_POS_X             = "last_pos_x"
	SETTING_LAST_POS_Y             = "last_pos_y"
	SETTING_LAST_POS_Z             = "last_pos_z"
	SETTING_TOTAL_LEGACY_COUNT     = "total_legacy_count"
	SETTING_PROCESSED_LEGACY_COUNT = "total_processed_legacy_count"
)

const getLastBlockQuery = `
select x,y,z,data,mtime
from blocks b
where b.x > ?
and b.y > ?
and b.z > ?
order by b.x asc, b.y asc, b.z asc, b.mtime asc
limit ?
`

func (a *Sqlite3Accessor) FindNextInitialBlocks(s settings.Settings, layers []*types.Layer, limit int) (*db.InitialBlocksResult, error) {
	result := &db.InitialBlocksResult{}

	blocks := make([]*db.Block, 0)
	lastpos_x := s.GetInt(SETTING_LAST_POS_X, -1 << (12 -1))
	lastpos_y := s.GetInt(SETTING_LAST_POS_Y, -1 << (12 -1))
	lastpos_z := s.GetInt(SETTING_LAST_POS_Z, -1 << (12 -1))

	processedcount := s.GetInt64(SETTING_PROCESSED_LEGACY_COUNT, 0)
	totallegacycount := s.GetInt64(SETTING_TOTAL_LEGACY_COUNT, -1)
	if totallegacycount == -1 {
		//Query from db
		totallegacycount, err := a.CountBlocks()

		if err != nil {
			panic(err)
		}

		s.SetInt64(SETTING_TOTAL_LEGACY_COUNT, int64(totallegacycount))
	}

	rows, err := a.db.Query(getLastBlockQuery, lastpos_x, lastpos_y, lastpos_z, limit)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		result.HasMore = true
		result.UnfilteredCount++

		var x int
		var y int
		var z int
		var data []byte
		var mtime int64

		err = rows.Scan(&x, &y, &z, &data, &mtime)
		if err != nil {
			return nil, err
		}

		if mtime > result.LastMtime {
			result.LastMtime = mtime
		}

		mb := convertRows(x, y, z, data, mtime)

		// new position
		lastpos_x = x
		lastpos_y = y
		lastpos_z = z

		blockcoordy := mb.Pos.Y
		currentlayer := types.FindLayerByY(layers, blockcoordy)

		if currentlayer != nil {
			blocks = append(blocks, mb)
		}
	}

	s.SetInt64(SETTING_PROCESSED_LEGACY_COUNT, int64(result.UnfilteredCount)+processedcount)

	result.Progress = float64(processedcount) / float64(totallegacycount)
	result.List = blocks

	//Save current positions of initial run
	s.SetInt(SETTING_LAST_POS_X, lastpos_x)
	s.SetInt(SETTING_LAST_POS_Y, lastpos_y)
	s.SetInt(SETTING_LAST_POS_Z, lastpos_z)

	return result, nil
}
