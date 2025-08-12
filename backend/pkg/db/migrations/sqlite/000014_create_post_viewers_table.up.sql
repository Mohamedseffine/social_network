CREATE TABLE post_viewers (
    post_id INTEGER NOT NULL,
    viewer_id INTEGER NOT NULL,
    PRIMARY KEY (post_id, viewer_id),
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE,
    FOREIGN KEY (viewer_id) REFERENCES users(id) ON DELETE CASCADE
);
