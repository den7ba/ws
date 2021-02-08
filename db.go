package main

import (
	"database/sql"
	"fmt"
)

//проверка подключения к бд
func checkDb() {
	flowPrintln("Check the database...")
	flowPrintln("▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓")
	fmt.Println("")
	err := db.Ping()
	if err != nil {
		flowPrintln("Проблемы соединения с базой данных. Сервер будет закрыт.")
		closeApp(err)
	}
}

func addMessageToDb(data JsonResponseMessage) (int, error) {

	tx, err := db.Begin()

	//получить id диалога, если его нет то создать todo: есть возможность создать 2 диалога!
	dialog, err := getDialog(data.From, data.Rcpt)

	if err != nil {
		if err == sql.ErrNoRows {
			dialog.Id = int(createDialog(data, tx))
		} else {
			if err != nil {
				return 0, err
			}
		}
	}

	// ***** ------------- *****
	res, err := tx.Exec("INSERT INTO messages SET `from` = ?, rcpt = ?, body = ?, dialog_id= ?",
		data.From, data.Rcpt, data.Body, dialog.Id)
	if err != nil {
		return 0, err
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// ***** ------------- *****

	_, err = tx.Exec("UPDATE dialogs SET last_message_id = ?, last_message_body = ?, viewed = 0, deleted = 0, last_id = ?, updated_at = DEFAULT WHERE id = ?",
		lastId, data.Body, data.From, dialog.Id)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("UPDATE users SET private = `private` + 1 WHERE id = ?",
		data.Rcpt)
	if err != nil {
		return 0, err
	}

	err = tx.Commit()

	return int(lastId), err
}

func getDialog(a, b int) (Dialog, error) {
	var m1, m2 int
	var dialog Dialog

	if a < b {
		m1, m2 = a, b
	} else {
		m1, m2 = b, a
	}

	row := db.QueryRow("SELECT id FROM dialogs WHERE member1=? and member2=?", m1, m2)
	return dialog, row.Scan(&dialog.Id)
}

func getUserId(id int) int {
	var rec User

	row := db.QueryRow("SELECT id FROM users WHERE id=?", id)

	_ = row.Scan(&rec.Id)

	return rec.Id
}

func getUser(id int) (User, error) {
	var rec User

	row := db.QueryRow("SELECT id, firstname, lastname, access, private FROM users WHERE id = ?", id)

	err := row.Scan(&rec.Id, &rec.Firstname, &rec.Lastname, &rec.Access, &rec.Private)

	return rec, err
}

//создание диалога (!!!продолжает транзакцию *sql.Tx)
func createDialog(data JsonResponseMessage, tx *sql.Tx) int64 {
	var m1, m2 int

	if data.From < data.Rcpt {
		m1, m2 = data.From, data.Rcpt
	} else {
		m1, m2 = data.Rcpt, data.From
	}

	res, err := tx.Exec("INSERT INTO dialogs SET member1 = ?, member2 = ?, viewed = 0, last_id = ?, last_message_body = ?",
		m1, m2, data.From, data.Body)
	fatal(err)

	lastId, err := res.LastInsertId()
	fatal(err)
	return lastId
}

//сделать сообщение просмотренным  	//data = message data
func viewMessage(data JsonResponseMessage, userId int, check bool) (JsonResponseMessage, error) {
	var rec JsonResponseMessage

	// создаем транзакцию
	tx, err := db.Begin()

	if check {
		//действительно ли сообщение этому пользователю и действительно ли оно непрочитанное
		row := tx.QueryRow("SELECT id, dialog_id, `from`, rcpt FROM messages WHERE id = ? AND `rcpt` = ? AND viewed = 0", data.Id, userId)
		err := row.Scan(&rec.Id, &rec.Dialog, &rec.From, &rec.Rcpt)
		if err != nil {
			fatal(err)
			return rec, err
		}
	}

	// диалог становится просмотренным в любом случае
	_, err = tx.Exec("UPDATE dialogs SET viewed = 1, updated_at = DEFAULT WHERE id = ?",
		data.Dialog)
	if err != nil {
		fatal(err)
		return rec, err
	}

	_, err = tx.Exec("UPDATE messages SET viewed = 1, updated_at = DEFAULT WHERE id = ?",
		data.Id)
	if err != nil {
		fatal(err)
		return rec, err
	}

	_, err = tx.Exec("UPDATE users SET private = `private` - 1 WHERE id = ?",
		userId)
	if err != nil {
		fatal(err)
		return rec, err
	}

	//коммит транзакции
	err = tx.Commit()
	if err != nil {
		fatal(err)
		return rec, err
	}

	fmt.Print("Сообщение №")
	fmt.Print(data.Id)
	fmt.Println(" просмотрено")

	if check {
		return rec, nil
	} else {
		return data, nil
	}
}

//прочитать последние limit сообщений
func readMessages(observerId, opponentId, first, limit int) (*sql.Rows, error) {
	dialog, err := getDialog(observerId, opponentId)

	var rows *sql.Rows

	if first == 0 {
		rows, err = db.Query("SELECT id, dialog_id, `from`, rcpt, body, viewed FROM messages WHERE dialog_id = ? ORDER BY id DESC LIMIT ?",
			dialog.Id, limit)
	} else {
		rows, err = db.Query("SELECT id, dialog_id, `from`, rcpt, body, viewed FROM messages WHERE dialog_id = ? AND id < ? ORDER BY id DESC LIMIT ?",
			dialog.Id, first, limit)
	}

	if err != nil {
		return nil, err
	}

	return rows, nil
}

/*
func readOne(id int) (Note, error) {
	var rec Note
	row := db.QueryRow("SELECT * FROM gotable WHERE id=?", id)
	return rec, row.Scan(&rec.Id, &rec.Name, &rec.Phone)
}
/*
func read(str string) ([]Note, error) {
	var rows *sql.Rows
	var err error
	if str != "" {
		rows, err = db.Query("SELECT * FROM gotable WHERE name LIKE ? ORDER BY id",
			"%"+str+"%")
	} else {
		rows, err = db.Query("SELECT * FROM gotable ORDER BY id")
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs = make([]Note, 0)
	var rec Note
	for rows.Next() {
		if err = rows.Scan(&rec.Id, &rec.Name, &rec.Phone); err != nil {
			return nil, err
		}
		rs = append(rs, rec)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return rs, nil
}


func insert(name, phone string) (sql.Result, error) {
	return db.Exec("INSERT INTO gotable (`name`, phone) VALUES (?, ?)",
		name, phone)
}

func remove(id int) (sql.Result, error) {
	return db.Exec("DELETE FROM gotable WHERE id=?", id)
}

func update(id int, name, phone string) (sql.Result, error) {
	return db.Exec("UPDATE gotable SET `name` = ?, phone = ? WHERE id=?",
		name, phone, id)
}
*/
