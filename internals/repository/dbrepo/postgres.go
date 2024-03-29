package dbrepo

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/TheDevCarnage/FortSmythesMotel/internals/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservations) (int, error){
	
	var newId int

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	statement := `insert into reservations (first_name, last_name, email, phone, start_date,
					end_date, room_id, created_at, updated_at)
					values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := m.DB.QueryRowContext(ctx, statement, 
				res.FirstName,
				res.LastName,
				res.Email, 
				res.Phone, 
				res.StartDate,
				res.EndDate,
				res.RoomID,
				time.Now(),
				time.Now(),
			).Scan(&newId)			
	if err != nil{
		return 0, err
	}

	return newId, nil
}


func (m *postgresDBRepo) InsertRoomRestriction(r models.RoomRestrictions) error{
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	statement := `insert into room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id)
					values ($1, $2, $3, $4, $5, $6, $7)`

	_, err := m.DB.ExecContext(ctx, statement, 
				r.StartDate,
				r.EndDate,
				r.RoomID,
				r.ReservationID,
				time.Now(),
				time.Now(),
				r.RestrictionID,
		)
		if err != nil {
			return err
		} 
	return nil
}

 

func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error){
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var numRows int

	statement := `select 
					count(id)
				  from 
				  	room_restrictions
				  where 
				  	room_id = $1 and
				    $2 < end_date and $3 > start_date;`
	
	row := m.DB.QueryRowContext(ctx, statement, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
			return false, err
		} 

	if numRows == 0{
		return true, nil
	}
	return false, nil
}



func (m *postgresDBRepo) SearchAvailabilityForAllRooms(start, end time.Time)([]models.Rooms, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	var rooms []models.Rooms

	statement := `select id, room_name 
				  from rooms r 
				  where r.id not in 
				  (select 
					id
				  from 
				  	room_restrictions rr
				  where 
				    $1 < end_date and $2 > start_date)`
	
	rows, err := m.DB.QueryContext(ctx, statement, start, end)
	if err != nil {
			return rooms, err
		} 
	for rows.Next(){
		var room models.Rooms
		err:= rows.Scan(
			&room.ID,
			&room.RoomName,
		)
		if err != nil{
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil{
		log.Fatal("Error scanning rows", err)
		return rooms, err
	} 
	return rooms, nil
} 


//GetRoomByID: get's room for given ID
func (m *postgresDBRepo) GetRoomByID(id int) (models.Rooms, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Rooms

	query:= `
	select id, room_name, created_at, updated_at from rooms where id=$1
	`
	row := m.DB.QueryRowContext(ctx, query, id)

	err:= row.Scan(
		&room.ID,
		&room.RoomName,
		&room.CreatedAt,
		&room.UpdatedAt,
	)
	if err!=nil{
		return room, err
	}

	return room, nil
}


//getUserByID: returns a user based on the given id
func (m *postgresDBRepo) GetUserByID(id int) (models.Users, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id first_name, last_name, email, password, access_level, created_at, updated_at 
			  from users where id=$1`

	row := m.DB.QueryRowContext(ctx, query, id)

	var u models.Users

	err := row.Scan(
		&u.ID, 
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil{
		return u, err
	}

	return u, nil
}



//UpdateUser: updates the user in the database
func (m *postgresDBRepo) UpdateUser(u *models.Users) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `
		update users set first_name=$1, last_name=$2, email=$3, access_level=$4, updated_at=$5`

	_, err := m.DB.ExecContext(ctx, query, 
		u.FirstName,
		u.LastName,
		u.Email,
		u.UpdatedAt,
		)
	if err != nil{
		return err
	}

	return nil
}


//AuthenticateUser: authenticates the user based on the email and password given in the payload
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error){
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	// query := ``

	row := m.DB.QueryRowContext(ctx, "select id, password from users where id=$1", email)
	
	err := row.Scan(
		&id,
		&hashedPassword,
	) 

	if err != nil{
		return id, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0,"", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return id, hashedPassword,nil


}