package db

import (
	"database/sql"
	"errors"
	"fmt"

	"social_network/internal/domain/models"
	"social_network/internal/domain/ports/repository"
)

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &UserRepositoryImpl{db: db}
}

func (r *UserRepositoryImpl) RegisterNewUser(user *models.User) error {
	query := `
        INSERT INTO users (
            email, password_hash, first_name, last_name, date_of_birth,
            avatar_path, user_name, about_me, privacy_status, gender
        )
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err := r.db.Exec(
		query,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.DateOfBirth,
		user.AvatarPath,
		user.UserName,
		user.AboutMe,
		user.PrivacyStatus,
		user.Gender,
	)
	return err
}

func (r *UserRepositoryImpl) GetUserByID(id int) (*models.User, error) {
	// fmt.Println(" user by ID:", id)
	query := `
		SELECT id, user_name, date_of_birth, gender, password_hash,
		       email, first_name, last_name, avatar_path, about_me, privacy_status, created_at
		FROM users WHERE id = ?
	`
	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.Id,
		&user.UserName,
		&user.DateOfBirth,
		&user.Gender,
		&user.Password,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.AvatarPath,
		&user.AboutMe,
		&user.PrivacyStatus,
		&user.CreatedAt,
	)
	// fmt.Println("Fetching user by ID:", user.Id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepositoryImpl) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, user_name, date_of_birth, gender, password_hash,
		       email, first_name, last_name, avatar_path, about_me, privacy_status, created_at
		FROM users WHERE email = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.Id,
		&user.UserName,
		&user.DateOfBirth,
		&user.Gender,
		&user.Password,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.AvatarPath,
		&user.AboutMe,
		&user.PrivacyStatus,
		&user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) UpdatePrivacyStatus(userID int, privacyStatus string) error {
	query := `
		UPDATE users
		SET privacy_status = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, privacyStatus, userID)
	return err
}

func (r *UserRepositoryImpl) SearchUsers(query string) ([]models.UserProfileDTO, error) {
	searchQuery := "%" + query + "%"
	sql := `
		SELECT id, user_name, first_name, last_name, avatar_path
		FROM users
		WHERE user_name LIKE ? OR first_name LIKE ? OR last_name LIKE ?
		LIMIT 20
	`
	rows, err := r.db.Query(sql, searchQuery, searchQuery, searchQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.UserProfileDTO
	for rows.Next() {
		var dto models.UserProfileDTO
		if err := rows.Scan(&dto.Id, &dto.UserName, &dto.FirstName, &dto.LastName, &dto.AvatarUrl); err != nil {
			return nil, err
		}
		users = append(users, dto)
	}
	return users, nil
}

func (r *UserRepositoryImpl) GetUserProfileByUsername(username string) (*models.UserProfileDTO, error) {
	query := `
		SELECT id, user_name, first_name, last_name, avatar_path, email, about_me, privacy_status, gender
		FROM users
		WHERE user_name = ?
	`
	var dto models.UserProfileDTO
	err := r.db.QueryRow(query, username).Scan(
		&dto.Id,
		&dto.UserName,
		&dto.FirstName,
		&dto.LastName,
		&dto.AvatarUrl,
		&dto.Email,
		&dto.AboutMe,
		&dto.PrivacyStatus,
		&dto.Gender,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &dto, nil
}

func (r *UserRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, date_of_birth, avatar_path, user_name, about_me, privacy_status, gender, created_at
		FROM users
		WHERE user_name = ?
		LIMIT 1
	`
	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.DateOfBirth,
		&user.AvatarPath,
		&user.UserName,
		&user.AboutMe,
		&user.PrivacyStatus,
		&user.Gender,
		&user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepositoryImpl) GetUserIDByUsername(username string) (int, error) {
	query := `
		SELECT id
		FROM users
		WHERE user_name = ?
	`
	var id int
	err := r.db.QueryRow(query, username).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("user not found")
		}
		return 0, err
	}
	return id, nil
}

// get chat users and groups for the chat:

// Get all users:
func (userRepo *UserRepositoryImpl) GetSortedUsersForChat(myID, offset, limit int) ([]*models.ChatUser, error) {
	query := `
	SELECT id, user_name, unread_count
	FROM (
		-- Users who have chatted with me
		SELECT 
			u.id, 
			u.user_name, 
			MAX(pm.created_at) AS last_message_time,
			(
				SELECT COUNT(*) 
				FROM private_messages 
				WHERE sender_id = u.id AND receiver_id = ? AND is_read = 0
			) AS unread_count
		FROM users u
		JOIN private_messages pm
			ON (u.id = pm.sender_id AND pm.receiver_id = ?) 
			OR (u.id = pm.receiver_id AND pm.sender_id = ?)
		WHERE u.id != ? AND u.privacy_status = 'public'
		GROUP BY u.id

		UNION ALL

		-- Users who have NOT chatted with me
		SELECT 
			u.id, 
			u.user_name, 
			NULL AS last_message_time,
			0 AS unread_count
		FROM users u
		WHERE u.id != ? 
		  AND u.privacy_status = 'public'
		  AND u.id NOT IN (
			SELECT 
				CASE 
					WHEN pm.sender_id = ? THEN pm.receiver_id
					ELSE pm.sender_id
				END
			FROM private_messages pm
			WHERE pm.sender_id = ? OR pm.receiver_id = ?
		)
	) AS all_users
	ORDER BY 
		last_message_time IS NULL,        -- Put users with no chat history last
		last_message_time DESC,           -- Recent chats first
		LOWER(user_name) ASC              -- Alphabetical tie-breaker
	LIMIT ? OFFSET ?;
	`

	rows, err := userRepo.db.Query(
		query,
		myID, myID, myID, myID, // for the first subquery
		myID, myID, myID, myID, // for the second subquery
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var users []*models.ChatUser
	for rows.Next() {
		chatUser := &models.ChatUser{}
		if err := rows.Scan(&chatUser.Id, &chatUser.NickName, &chatUser.UnreadCount); err != nil {
			return nil, err
		}

		chatUser.IsOnline = false

		users = append(users, chatUser)
	}
for i, v := range users {
	fmt.Printf("the %dth user is %v\n", i, v)
}
	return users, nil
}

