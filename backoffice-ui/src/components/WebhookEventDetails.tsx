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
import { api } from '../services/api';
import { WebhookEvent } from '../types/webhook';
import { useNotification } from '../context/NotificationContext';

export const WebhookEventDetails: React.FC = () => {
  const [webhookEvent, setWebhookEvent] = useState<WebhookEvent | null>(null);
  const [loading, setLoading] = useState(true);
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { showError } = useNotification();

  useEffect(() => {
    if (id) {
      loadWebhookEvent(id);
    }
  }, [id]);

  const loadWebhookEvent = async (eventId: string) => {
    try {
      setLoading(true);
      const data = await api.getWebhookEventById(eventId);
      setWebhookEvent(data);
    } catch (error) {
      showError('Failed to load webhook event details');
      console.error('Failed to load webhook event:', error);
    } finally {
      setLoading(false);
    }
  };

  const getEventTypeColor = (eventType: string) => {
    if (eventType.includes('created')) return 'success';
    if (eventType.includes('updated')) return 'info';
    if (eventType.includes('deleted')) return 'error';
    return 'default';
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

  if (!webhookEvent) {
    return (
      <Box>
        <Alert severity="error">Webhook event not found</Alert>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/webhook-events')}
          sx={{ mt: 2 }}
        >
          Back to Webhook Events
        </Button>
      </Box>
    );
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" mb={3}>
        <Button
          startIcon={<ArrowBackIcon />}
          onClick={() => navigate('/webhook-events')}
          sx={{ mr: 2 }}
        >
          Back
        </Button>
        <Typography variant="h4" component="h1">
          Webhook Event Details
        </Typography>
      </Box>

      <Grid container spacing={3}>
        <Grid item xs={12}>
          <Card>
            <CardContent>
              <Box display="flex" justifyContent="space-between" alignItems="center" mb={2}>
                <Typography variant="h5" component="h2">
                  {webhookEvent.jira_issue_key}
                </Typography>
                <Chip
                  label={webhookEvent.event_type}
                  color={getEventTypeColor(webhookEvent.event_type)}
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
                    {webhookEvent.jira_issue_id || 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    JIRA Project Key
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {webhookEvent.jira_project_key}
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Summary
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {webhookEvent.summary}
                  </Typography>
                </Grid>
                <Grid item xs={12}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Description
                  </Typography>
                  <Typography variant="body1" gutterBottom sx={{ whiteSpace: 'pre-wrap' }}>
                    {webhookEvent.description || 'No description'}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Status
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {webhookEvent.status || 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Previous Status
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {webhookEvent.previous_status || 'N/A'}
                  </Typography>
                </Grid>
                <Grid item xs={12} sm={6}>
                  <Typography variant="subtitle2" color="text.secondary">
                    Received At
                  </Typography>
                  <Typography variant="body1" gutterBottom>
                    {formatDate(webhookEvent.received_at)}
                  </Typography>
                </Grid>
                {webhookEvent.processed_at && (
                  <Grid item xs={12} sm={6}>
                    <Typography variant="subtitle2" color="text.secondary">
                      Processed At
                    </Typography>
                    <Typography variant="body1" gutterBottom>
                      {formatDate(webhookEvent.processed_at)}
                    </Typography>
                  </Grid>
                )}
              </Grid>
            </CardContent>
          </Card>
        </Grid>

        {webhookEvent.raw_payload && (
          <Grid item xs={12}>
            <Card>
              <CardContent>
                <Typography variant="h6" gutterBottom>
                  Raw Webhook Payload
                </Typography>
                <Paper
                  variant="outlined"
                  sx={{
                    p: 2,
                    backgroundColor: 'grey.50',
                    maxHeight: 500,
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
                    {JSON.stringify(webhookEvent.raw_payload, null, 2)}
                  </Typography>
                </Paper>
              </CardContent>
            </Card>
          </Grid>
        )}
      </Grid>
    </Box>
  );
};
