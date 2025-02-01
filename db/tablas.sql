CREATE TABLE Posts (
    ID INT AUTO_INCREMENT,
    Title VARCHAR(100),
    Description TEXT,
    Content LONGTEXT,
    AuthorID INT,
    CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
    Category VARCHAR(50),
    PRIMARY KEY (ID)
);

CREATE TABLE Users (
    ID INT AUTO_INCREMENT,
    Username VARCHAR(100),
    Password CHAR(64),
    PRIMARY KEY (ID)
);