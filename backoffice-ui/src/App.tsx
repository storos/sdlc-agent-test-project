import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import { theme } from './theme/theme';
import { NotificationProvider } from './context/NotificationContext';
import { Layout } from './components/Layout';
import { ProjectList } from './components/ProjectList';
import { ProjectForm } from './components/ProjectForm';
import { RepositoryList } from './components/RepositoryList';
import { DevelopmentList } from './components/DevelopmentList';
import { DevelopmentDetails } from './components/DevelopmentDetails';

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <NotificationProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<Navigate to="/projects" replace />} />
              <Route path="projects" element={<ProjectList />} />
              <Route path="projects/new" element={<ProjectForm />} />
              <Route path="projects/:id/edit" element={<ProjectForm />} />
              <Route path="projects/:projectId/repositories" element={<RepositoryList />} />
              <Route path="developments" element={<DevelopmentList />} />
              <Route path="developments/:id" element={<DevelopmentDetails />} />
            </Route>
          </Routes>
        </BrowserRouter>
      </NotificationProvider>
    </ThemeProvider>
  );
}

export default App;
