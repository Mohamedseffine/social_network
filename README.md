# Social Network

## Overview

This project is a Facebook-like social network application built with a
**Next.js frontend** and a **Go backend**. It supports core social
features such as user authentication, posts, groups, messaging, and
real-time notifications.

------------------------------------------------------------------------

## Tech Stack

-   **Frontend:** Next.js (React, JavaScript, CSS)
-   **Backend:** Go (standard library)
-   **Database:** SQLite
-   **Real-time Communication:** WebSockets (Gorilla WebSocket)
-   **Containerization:** Docker

------------------------------------------------------------------------

## Features

### Authentication

-   User registration and login
-   Sessions and cookies for persistent authentication
-   Optional profile fields (avatar, nickname, bio)

### Profiles

-   Public and private profiles
-   User info, posts, followers, and activity
-   Toggle profile visibility

### Followers

-   Follow/unfollow users
-   Follow requests for private profiles
-   Automatic follow for public profiles

### Posts

-   Create posts and comments
-   Attach images (JPEG, PNG, GIF)
-   Privacy levels:
    -   Public
    -   Followers only
    -   Selected users

### Groups

-   Create and manage groups
-   Invite users and handle join requests
-   Group posts and comments
-   Events with RSVP (Going / Not Going)

### Chat

-   Private messaging between users
-   Real-time communication via WebSockets
-   Group chat rooms
-   Emoji support

### Notifications

-   Follow requests
-   Group invitations
-   Group join requests (for admins)
-   Event creation alerts

------------------------------------------------------------------------

## Backend Architecture

-   HTTP server handling API requests
-   Middleware for authentication and request handling
-   SQLite database for persistent storage
-   File system for image storage
-   WebSocket server for real-time messaging

------------------------------------------------------------------------

## Database & Migrations

Migrations are stored in:

    backend/pkg/db/migrations/sqlite

Example:

    000001_create_users_table.up.sql
    000001_create_users_table.down.sql

-   Managed using migration tools (e.g., golang-migrate)
-   Automatically applied on application startup

------------------------------------------------------------------------

## Docker Setup

### Backend Container

-   Runs Go server
-   Handles API and database interaction
-   Exposes backend port

### Frontend Container

-   Runs Next.js app
-   Serves client-side application
-   Communicates with backend via HTTP

------------------------------------------------------------------------

## Installation

### Prerequisites

-   Docker
-   Go (if running locally without Docker)
-   Node.js (for frontend development)

### Run with Docker

``` bash
docker-compose up --build
```

### Run Backend Locally

``` bash
go run backend/server.go
```

### Run Frontend Locally

``` bash
cd frontend
npm install
npm run dev
```

------------------------------------------------------------------------

## Project Structure

    backend/
      pkg/
        db/
          migrations/
            sqlite/
          sqlite/
    frontend/
      pages/
      components/

------------------------------------------------------------------------

## Learning Objectives

-   Authentication with sessions and cookies
-   Database design and SQL (SQLite)
-   Real-time systems using WebSockets
-   Docker containerization
-   Full-stack application architecture

------------------------------------------------------------------------

## Notes

-   Images are stored on disk with paths saved in the database
-   WebSockets are used for chat and real-time updates
-   The application follows a modular backend structure

------------------------------------------------------------------------

## License

This project is for educational purposes.
