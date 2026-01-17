import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  Grid,
  Paper,
  Typography,
  Alert,
} from '@mui/material';
import { useNavigate, useParams } from 'react-router-dom';
import ArrowBackIcon from '@mui/icons-material/ArrowBack';
import OpenInNewIcon from '@mui/icons-material/OpenInNew';
import { api } from '../services/api';
import { Development } from '../types/development';
import { useNotification } from '../context/NotificationContext';

export const DevelopmentDetails: React.FC = () => {
  const [development, setDevelopment] = useState<Development | null>(null);
  const [loading, setLoading] = useState(true);
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showError } = useNotification();

  useEffect(() => {
    if (id) {
      loadDevelopment(id);
    }
  }, [id]);

  const loadDevelopment = async (devId: string) => {
    try {
      setLoading(true);
      const data = await api.getDevelopmentById(devId);
      setDevelopment(data);
    } catch (error) {
      showError('Failed to load development details');
      console.error('Failed to load development:', error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'failed':
        return 'error';
      case 'ready':
        return 'warning';
      default:
        return 'default';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (loading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  if (!development) {
    return (
      <Box>
        <Alert severity="error">Development not found</Alert>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/developments')}
          sx={{ mt: 2 }}
        >
          Back to Developments
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" mb={3}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/developments')}
          sx={{ mr: 2 }}
        >
          Back
        </Button>
        <Typography variant="h4" component="h1">
          Development Details
        </Typography>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h5" component="h2">
                  {development.jira_issue_key}
                </Typography>
                <Chip
                  label={development.status}
                  color={getStatusColor(development.status)}
                  size="medium"
                />
              </Box>
              <Divider sx={{ mb: 2 }} />
              <Grid container spacing={2}>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    JIRA Issue ID
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {development.jira_issue_id}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    JIRA Project Key
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {development.jira_project_key}
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Repository URL
                  </Typography>
                  <Typography variant="body1" gutterBottom sx={{ wordBreak: 'break-all' }}>
                    {development.repository_url}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Branch Name
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {development.branch_name}
                  </Typography>
                </Grid>
                {development.pr_mr_url && (
                  <Grid item xs={12} sm={6}>
                    <Typography variant="subtitle2" color="text.secondary">
                      Pull Request / Merge Request
                    </Typography>
                    <Button
                      href={development.pr_mr_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      endIcon={<OpenInNewIcon />}
                      variant="outlined"
                      size="small"
                      sx={{ mt: 0.5 }}
                    >
                      View PR/MR
                    </Button>
                  </Grid>
                )}
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Created At
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {formatDate(development.created_at)}
                  </Typography>
                </Grid>
                {development.completed_at && (
                  <Grid item xs={12} sm={6}>
                    <Typography variant="subtitle2" color="text.secondary">
                      Completed At
                    </Typography>
                    <Typography variant="body1" gutterBottom>
                      {formatDate(development.completed_at)}
                    </Typography>
                  </Grid>
                )}
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {development.development_details && (
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Development Details
                </Typography>
                <Paper
                  variant="outlined"
                  sx={{
                    p: 2,
                    backgroundColor: 'grey.50',
                    maxHeight: 400,
                    overflow: 'auto',
                  }}
                >
                  <Typography
                    variant="body2"
                    component="pre"
                    sx={{
                      whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                      fontFamily: 'monospace',
                      fontSize: '0.875rem',
                    }}
                  >
                    {development.development_details}
                  </Typography>
                </Paper>
              </CardContent>
            </Card>
          </Grid>
        )}

        {development.error_message && (
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom color="error">
                  Error Message
                </Typography>
                <Alert severity="error">
                  <Typography variant="body2" component="pre" sx={{ whiteSpace: 'pre-wrap' }}>
                    {development.error_message}
                  </Typography>
                </Alert>
              </CardContent>
            </Card>
          </Grid>
        )}
      </Grid>
    </Box>
  );
};
