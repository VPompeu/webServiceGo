CREATE TABLE users (
    ID SERIAL PRIMARY KEY,
    Name VARCHAR(255),
    Email VARCHAR(255),
    Password VARCHAR(255),
    Phone VARCHAR(20),
    Birthday VARCHAR(255),
    City VARCHAR(255),
    State VARCHAR(255),
    Country VARCHAR(255),
    License BOOLEAN
);
ALTER TABLE users ADD CONSTRAINT unique_email UNIQUE (email);

CREATE TABLE user_notes (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    note TEXT NOT NULL,
    note_date VARCHAR(10) NOT NULL,
    CONSTRAINT fk_user
      FOREIGN KEY(user_id) 
	  REFERENCES users(id),
    UNIQUE (user_id, note_date)
);

CREATE TABLE global_notes (
    id SERIAL PRIMARY KEY,
    note TEXT NOT NULL,
    note_date VARCHAR(10) NOT NULL UNIQUE
);

CREATE TABLE password_reset_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    token TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

UPDATE users
SET license = false
WHERE id = 1;

INSERT INTO `users` (`ID`, `Name`, `Email`, `Password`, `Phone`, `Birthday`, `City`, `State`, `Country`, `License`) VALUES (1, "Victor Pompeu", "pompeu.dev@gmail.com", "$2a$10$7uzd8Dx7MwIdxU3ZBoK.jePYUVlPHjAnkrb3LxwVbUlA1/RcCNHOG", "51996368303", "04/03/2001", "Tramandaí", "RS", "Brasil", 1);