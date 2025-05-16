# deeL

A minimalist RSS feed reader with a clean interface and dark mode support.
![deeL Logo](static/images/deeL-logo.png)

## Features

- Add and manage RSS feeds
- Clean, responsive interface
- Dark/Light theme toggle
- Auto-refresh feeds
- Mobile-friendly design

## Setup

1. Clone the repository:
```bash
git clone https://github.com/yourusername/deeL.git
cd deeL
```

2. Install dependencies and build:
```bash
go mod tidy
make build
```

3. Run the application:
```bash
make run
```

The application will be available at `http://localhost:8080` by default.

## Project Structure

```
├── cmd
│   └── server         # Entry point for the application
├── internal
│   ├── database       # Database operations
│   ├── feeds          # Feed processing and management
│   ├── handlers       # HTTP handlers
│   ├── models         # Data structures
│   └── utils          # Utility functions
├── static             # Static assets (CSS, JS, images)
└── templates          # HTML templates
```

## Usage

- Click "Add a new RSS feed" to add RSS feed URLs
- Use the theme toggle in the top right to switch between light and dark modes
- Click "Refresh All Feeds" to update your feed content
- Click the hamburger menu on mobile to show/hide the sidebar

## License

MIT License
