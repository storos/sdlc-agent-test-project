import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  TextField,
  Typography,
} from '@mui/material';
import { useNavigate, useParams } from 'react-router-dom';
import { api } from '../services/api';
import { useNotification } from '../context/NotificationContext';
import type { CreateProjectRequest, UpdateProjectRequest } from '../types/project';

export const ProjectForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { showSuccess, showError } = useNotification();
  const isEdit = Boolean(id);

  const [formData, setFormData] = useState<CreateProjectRequest>({
    name: '',
    description: '',
    scope: '',
    jira_project_key: '',
    jira_project_name: '',
    jira_project_url: '',
  });
  const [loading, setLoading] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (isEdit && id) {
      loadProject(id);
    }
  }, [isEdit, id]);

  const loadProject = async (projectId: string) => {
    try {
      setLoading(true);
      const project = await api.getProjectById(projectId);
      setFormData({
        name: project.name,
        description: project.description,
        scope: project.scope,
        jira_project_key: project.jira_project_key,
        jira_project_name: project.jira_project_name,
        jira_project_url: project.jira_project_url,
      });
    } catch (error) {
      showError('Failed to load project');
      console.error('Failed to load project:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (field: keyof CreateProjectRequest) => (
    e: React.ChangeEvent<HTMLInputElement>
  ) => {
    setFormData({ ...formData, [field]: e.target.value });
    if (errors[field]) {
      setErrors({ ...errors, [field]: '' });
    }
  };

  const validate = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    if (!formData.description.trim()) {
      newErrors.description = 'Description is required';
    }
    if (!formData.scope.trim()) {
      newErrors.scope = 'Scope is required';
    }
    if (!formData.jira_project_key.trim()) {
      newErrors.jira_project_key = 'JIRA Project Key is required';
    }
    if (!formData.jira_project_name.trim()) {
      newErrors.jira_project_name = 'JIRA Project Name is required';
    }
    if (!formData.jira_project_url.trim()) {
      newErrors.jira_project_url = 'JIRA Project URL is required';
    } else if (!formData.jira_project_url.match(/^https?:\/\/.+/)) {
      newErrors.jira_project_url = 'Invalid URL format';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    try {
      setLoading(true);
      if (isEdit && id) {
        const updateData: UpdateProjectRequest = formData;
        await api.updateProject(id, updateData);
        showSuccess('Project updated successfully');
      } else {
        await api.createProject(formData);
        showSuccess('Project created successfully');
      }
      navigate('/');
    } catch (error) {
      showError('Failed to save project');
      console.error('Failed to save project:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3 }}>
        {isEdit ? 'Edit Project' : 'New Project'}
      </Typography>

      <Card>
        <CardContent>
          <Box component="form" onSubmit={handleSubmit}>
            <TextField
              fullWidth
              label="Name"
              value={formData.name}
              onChange={handleChange('name')}
              error={Boolean(errors.name)}
              helperText={errors.name}
              disabled={loading}
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label="Description"
              value={formData.description}
              onChange={handleChange('description')}
              error={Boolean(errors.description)}
              helperText={errors.description}
              disabled={loading}
              multiline
              rows={3}
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label="Scope"
              value={formData.scope}
              onChange={handleChange('scope')}
              error={Boolean(errors.scope)}
              helperText={errors.scope}
              disabled={loading}
              multiline
              rows={3}
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label="JIRA Project Key"
              value={formData.jira_project_key}
              onChange={handleChange('jira_project_key')}
              error={Boolean(errors.jira_project_key)}
              helperText={errors.jira_project_key}
              disabled={loading || isEdit}
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label="JIRA Project Name"
              value={formData.jira_project_name}
              onChange={handleChange('jira_project_name')}
              error={Boolean(errors.jira_project_name)}
              helperText={errors.jira_project_name}
              disabled={loading}
              sx={{ mb: 2 }}
            />

            <TextField
              fullWidth
              label="JIRA Project URL"
              value={formData.jira_project_url}
              onChange={handleChange('jira_project_url')}
              error={Boolean(errors.jira_project_url)}
              helperText={errors.jira_project_url}
              disabled={loading}
              sx={{ mb: 3 }}
            />

            <Box sx={{ display: 'flex', gap: 2 }}>
              <Button
                variant="contained"
                type="submit"
                disabled={loading}
              >
                {isEdit ? 'Update' : 'Create'}
              </Button>
              <Button
                variant="outlined"
                onClick={() => navigate('/')}
                disabled={loading}
              >
                Cancel
              </Button>
            </Box>
          </Box>
        </CardContent>
      </Card>
    </Box>
  );
};
