import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Storage as StorageIcon,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { useNotification } from '../context/NotificationContext';
import type { Project } from '../types/project';

export const ProjectList: React.FC = () => {
  const navigate = useNavigate();
  const { showSuccess, showError } = useNotification();
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [projectToDelete, setProjectToDelete] = useState<Project | null>(null);

  const loadProjects = async () => {
    try {
      setLoading(true);
      const data = await api.getAllProjects();
      setProjects(data);
    } catch (error) {
      showError('Failed to load projects');
      console.error('Failed to load projects:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProjects();
  }, []);

  const handleDelete = async () => {
    if (!projectToDelete) return;

    try {
      await api.deleteProject(projectToDelete.id);
      await loadProjects();
      setDeleteDialogOpen(false);
      setProjectToDelete(null);
      showSuccess('Project deleted successfully');
    } catch (error) {
      showError('Failed to delete project');
      console.error('Failed to delete project:', error);
    }
  };

  const openDeleteDialog = (project: Project) => {
    setProjectToDelete(project);
    setDeleteDialogOpen(true);
  };

  return (
    <Box>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
        <Typography variant="h4">Projects</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => navigate('/projects/new')}
        >
          New Project
        </Button>
      </Box>

      <Card>
        <CardContent>
          {loading ? (
            <Typography>Loading...</Typography>
          ) : projects.length === 0 ? (
            <Typography>No projects found. Create your first project!</Typography>
          ) : (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Name</TableCell>
                    <TableCell>JIRA Project Key</TableCell>
                    <TableCell>JIRA Project Name</TableCell>
                    <TableCell>Repositories</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {projects.map((project) => (
                    <TableRow key={project.id}>
                      <TableCell>{project.name}</TableCell>
                      <TableCell>{project.jira_project_key}</TableCell>
                      <TableCell>{project.jira_project_name}</TableCell>
                      <TableCell>{project.repositories?.length || 0}</TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() => navigate(`/projects/${project.id}/repositories`)}
                          title="Manage Repositories"
                        >
                          <StorageIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => navigate(`/projects/${project.id}/edit`)}
                          title="Edit"
                        >
                          <EditIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => openDeleteDialog(project)}
                          title="Delete"
                        >
                          <DeleteIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Delete Project</DialogTitle>
        <DialogContent>
          Are you sure you want to delete the project "{projectToDelete?.name}"? This action cannot be undone.
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDeleteDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleDelete} color="error" variant="contained">
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
