package roach

import (
	"database/sql"
	"github.com/tomogoma/crdb"
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/usersms/pkg/rating"
)

var allRatingCols = ColDesc(ColID, ColForSection, ColForUserID, ColByUserID,
	ColRating, ColComment, ColCreated, ColLastUpdated)

func (r *Roach) SaveRating(rt rating.Rating) error {
	if err := r.InitDBIfNot(); err != nil {
		return err
	}
	cols := ColDesc(ColID, ColForSection, ColForUserID, ColByUserID, ColRating,
		ColComment, ColCreated, ColLastUpdated)
	q := `INSERT INTO ` + TblRatings + `(` + cols + `)
			VALUES $1, $2, $3, $4, $5, $6, $7, $8`
	res, err := r.db.Exec(q, rt.ID, rt.ForSection, rt.ForUserID, rt.ByUserID,
		rt.Rating, rt.Comment, rt.Created, rt.LastUpdated)
	return checkRowsAffected(res, err, 1)
}

func (r *Roach) Rating(byUserID, forSection, forUserID string) (*rating.Rating, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}

	q := `
		SELECT ` + allRatingCols + ` FROM ` + TblRatings + `
			WHERE ` + ColByUserID + `=$1
				AND ` + ColForSection + `=$2
				AND ` + ColForUserID + `=$3
	`

	rt, err := scanRating(r.db.QueryRow(q, byUserID, forSection, forUserID))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFound("no rating found for filter")
		}
		return nil, err
	}

	return rt, nil
}

func (r *Roach) Ratings(f rating.Filter) ([]rating.Rating, error) {
	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}

	whereOp := "AND"
	var where string
	var args []interface{}
	where, args = crdb.ConcatWhereClause(f.ForSection, ColForUserID, where, whereOp, args)
	where, args = crdb.ConcatWhereClause(f.ForUserID, ColForUserID, where, whereOp, args)
	where, args = crdb.ConcatWhereClause(f.ByUserID, ColByUserID, where, whereOp, args)

	limit, args := crdb.Pagination(f.Offset, int64(f.Count), args)

	q := `SELECT ` + allRatingCols + ` FROM ` + TblRatings + ` WHERE ` + where + ` ` + limit
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rts []rating.Rating
	for rows.Next() {
		rt, err := scanRating(rows)
		if err != nil {
			return nil, errors.Newf("scan rating from row: %v", err)
		}
		rts = append(rts, *rt)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterate result set: %v", err)
	}

	if len(rts) == 0 {
		return nil, errors.NewNotFound("no rating found for filter")
	}

	return rts, nil
}

func (r *Roach) AverageUserRatings(offset int64, count int32) ([]rating.AverageUser, error) {

	if err := r.InitDBIfNot(); err != nil {
		return nil, err
	}

	cols := ColDesc(ColForUserID, "AVG("+ColRating+")", "COUNT("+ColRating+")")
	limit, args := crdb.Pagination(offset, int64(count), []interface{}{})
	q := `SELECT ` + cols + ` FROM ` + TblRatings + ` GROUP BY ` + ColForUserID + ` ` + limit

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trs []rating.AverageUser
	for rows.Next() {
		tr := rating.AverageUser{}
		err := rows.Scan(&tr.UserID, &tr.Rating, &tr.NumRaters)
		if err != nil {
			return nil, errors.Newf("scan row in result set: %v", err)
		}
		trs = append(trs, tr)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Newf("iterating result set: %v", err)
	}

	if len(trs) == 0 {
		return nil, errors.NewNotFound("no user ratings found in range")
	}

	return trs, nil
}

// scanUser extracts a rating from s or returns an error if reported by s.
// The column order for s must be same order as allRatingCols variable.
func scanRating(s multiScanner) (*rating.Rating, error) {
	rt := &rating.Rating{}
	comment := sql.NullString{}
	err := s.Scan(&rt.ID, &rt.ForSection, &rt.ForUserID, &rt.ByUserID,
		&rt.Rating, &comment, &rt.Created, &rt.LastUpdated)
	if err != nil {
		return nil, err
	}
	rt.Comment = comment.String
	return rt, nil
}
