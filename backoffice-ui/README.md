# Backoffice UI

Web interface for managing SDLC AI Agents configuration.

## Features

- Project Management (Create, Read, Update, Delete)
- Repository Management per Project
- JIRA Project Integration
- Material-UI Components
- Real-time Notifications
- Form Validation

## Tech Stack

- React 18
- TypeScript
- Vite
- Material-UI (MUI)
- React Router
- Axios

## Development

### Prerequisites
- Node.js 18+
- Configuration API running on port 8081

### Setup

```bash
# Install dependencies
npm install

# Set environment variables (optional)
cp .env.example .env

# Start development server
npm run dev
```

The application will be available at http://localhost:3000

### Environment Variables

- `VITE_API_URL`: Configuration API URL (default: http://localhost:8081)

### Build

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

## Docker

### Build Image

```bash
docker build -t backoffice-ui .
```

### Run Container

```bash
docker run -p 3000:80 backoffice-ui
```

## Usage

### Projects Page

- View all configured projects
- Create new project
- Edit existing project
- Delete project
- Navigate to repository management

### Project Form

Required fields:
- Name
- Description
- Scope
- JIRA Project Key
- JIRA Project Name
- JIRA Project URL

### Repository Management

- View repositories for a project
- Add new repository
- Edit repository details
- Delete repository

## API Integration

The UI communicates with the Configuration API:

- `GET /api/projects` - List all projects
- `GET /api/projects/:id` - Get project by ID
- `POST /api/projects` - Create project
- `PUT /api/projects/:id` - Update project
- `DELETE /api/projects/:id` - Delete project
- `GET /api/projects/:id/repositories` - List repositories
- `POST /api/projects/:id/repositories` - Add repository
- `PUT /api/repositories/:id` - Update repository
- `DELETE /api/repositories/:id` - Delete repository
