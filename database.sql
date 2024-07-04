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

INSERT INTO `users` (`ID`, `Name`, `Email`, `Password`, `Phone`, `Birthday`, `City`, `State`, `Country`, `License`) VALUES (1, "Victor Pompeu", "pompeu.dev@gmail.com", "Vd323627", "51996368303", "04/03/2001", "Tramandaí", "RS", "Brasil", 1);