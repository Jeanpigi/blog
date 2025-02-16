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

CREATE TABLE Visits (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ip VARCHAR(45),
    user_agent TEXT,
    page VARCHAR(255),
    country VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    latitude DECIMAL(9,6),
    longitude DECIMAL(9,6)
);


