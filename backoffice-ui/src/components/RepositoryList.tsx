import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  IconButton,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  TextField,
  Typography,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  ArrowBack as ArrowBackIcon,
} from '@mui/icons-material';
import { useNavigate, useParams } from 'react-router-dom';
import { api } from '../services/api';
import { useNotification } from '../context/NotificationContext';
import type { Repository, AddRepositoryRequest, UpdateRepositoryRequest, Project } from '../types/project';

export const RepositoryList: React.FC = () => {
  const navigate = useNavigate();
  const { projectId } = useParams<{ projectId: string }>();
  const { showSuccess, showError } = useNotification();

  const [project, setProject] = useState<Project | null>(null);
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [loading, setLoading] = useState(true);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRepo, setEditingRepo] = useState<Repository | null>(null);
  const [formData, setFormData] = useState<AddRepositoryRequest>({
    url: '',
    description: '',
    git_access_token: '',
  });
  const [errors, setErrors] = useState<Record<string, string>>({});

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [repoToDelete, setRepoToDelete] = useState<Repository | null>(null);

  useEffect(() => {
    if (projectId) {
      loadData(projectId);
    }
  }, [projectId]);

  const loadData = async (id: string) => {
    try {
      setLoading(true);
      const [projectData, reposData] = await Promise.all([
        api.getProjectById(id),
        api.getRepositories(id),
      ]);
      setProject(projectData);
      setRepositories(reposData);
    } catch (error) {
      showError('Failed to load data');
      console.error('Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const openAddDialog = () => {
    setEditingRepo(null);
    setFormData({
      url: '',
      description: '',
      git_access_token: '',
      base_branch: 'main',
    });
    setErrors({});
    setDialogOpen(true);
  };

  const openEditDialog = (repo: Repository) => {
    setEditingRepo(repo);
    setFormData({
      url: repo.url,
      description: repo.description,
      git_access_token: repo.git_access_token,
      base_branch: repo.base_branch || 'main',
    });
    setErrors({});
    setDialogOpen(true);
  };

  const handleChange = (field: keyof AddRepositoryRequest) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    setFormData({ ...formData, [field]: e.target.value });
    if (errors[field]) {
      setErrors({ ...errors, [field]: '' });
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.url.trim()) {
      newErrors.url = 'URL is required';
    } else if (!formData.url.match(/^https?:\/\/.+/)) {
      newErrors.url = 'Invalid URL format';
    }
    if (!formData.description.trim()) {
      newErrors.description = 'Description is required';
    }
    if (!formData.git_access_token.trim()) {
      newErrors.git_access_token = 'Git Access Token is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSave = async () => {
    if (!validate() || !projectId) return;

    try {
      if (editingRepo) {
        const updateData: UpdateRepositoryRequest = formData;
        await api.updateRepository(projectId, editingRepo.repository_id, updateData);
        showSuccess('Repository updated successfully');
      } else {
        await api.addRepository(projectId, formData);
        showSuccess('Repository added successfully');
      }
      await loadData(projectId);
      setDialogOpen(false);
    } catch (error) {
      showError('Failed to save repository');
      console.error('Failed to save repository:', error);
    }
  };

  const handleDelete = async () => {
    if (!repoToDelete || !projectId) return;

    try {
      await api.deleteRepository(projectId, repoToDelete.repository_id);
      await loadData(projectId);
      setDeleteDialogOpen(false);
      setRepoToDelete(null);
      showSuccess('Repository deleted successfully');
    } catch (error) {
      showError('Failed to delete repository');
      console.error('Failed to delete repository:', error);
    }
  };

  const openDeleteDialog = (repo: Repository) => {
    setRepoToDelete(repo);
    setDeleteDialogOpen(true);
  };

  if (loading) {
    return <Typography>Loading...</Typography>;
  }

  if (!project) {
    return <Typography>Project not found</Typography>;
  }

  return (
    <Box>
      <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
        <IconButton onClick={() => navigate('/')} sx={{ mr: 2 }}>
          <ArrowBackIcon />
        </IconButton>
        <Typography variant="h4" sx={{ flexGrow: 1 }}>
          Repositories - {project.name}
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={openAddDialog}
        >
          Add Repository
        </Button>
      </Box>

      <Card>
        <CardContent>
          {repositories.length === 0 ? (
            <Typography>No repositories found. Add your first repository!</Typography>
          ) : (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>URL</TableCell>
                    <TableCell>Description</TableCell>
                    <TableCell>Base Branch</TableCell>
                    <TableCell align="right">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {repositories.map((repo) => (
                    <TableRow key={repo.repository_id}>
                      <TableCell>{repo.url}</TableCell>
                      <TableCell>{repo.description}</TableCell>
                      <TableCell>{repo.base_branch || 'main'}</TableCell>
                      <TableCell align="right">
                        <IconButton
                          size="small"
                          onClick={() => openEditDialog(repo)}
                          title="Edit"
                        >
                          <EditIcon />
                        </IconButton>
                        <IconButton
                          size="small"
                          onClick={() => openDeleteDialog(repo)}
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

      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>{editingRepo ? 'Edit Repository' : 'Add Repository'}</DialogTitle>
        <DialogContent>
          <Box sx={{ pt: 2 }}>
            <TextField
              fullWidth
              label="Repository URL"
              value={formData.url}
              onChange={handleChange('url')}
              error={Boolean(errors.url)}
              helperText={errors.url}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={handleChange('description')}
              error={Boolean(errors.description)}
              helperText={errors.description}
              multiline
              rows={2}
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Git Access Token"
              value={formData.git_access_token}
              onChange={handleChange('git_access_token')}
              error={Boolean(errors.git_access_token)}
              helperText={errors.git_access_token}
              type="password"
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Base Branch"
              value={formData.base_branch || 'main'}
              onChange={handleChange('base_branch')}
              error={Boolean(errors.base_branch)}
              helperText={errors.base_branch || 'Default branch for pull requests (e.g., main, master, develop)'}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>Cancel</Button>
          <Button onClick={handleSave} variant="contained">
            {editingRepo ? 'Update' : 'Add'}
          </Button>
        </DialogActions>
      </Dialog>

      <Dialog open={deleteDialogOpen} onClose={() => setDeleteDialogOpen(false)}>
        <DialogTitle>Delete Repository</DialogTitle>
        <DialogContent>
          Are you sure you want to delete the repository "{repoToDelete?.url}"? This action cannot be undone.
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
