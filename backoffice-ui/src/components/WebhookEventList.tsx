import React, { useEffect, useState } from 'react';
import {
  Box,
  Button,
  Chip,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Typography,
  CircularProgress,
  Alert,
} from '@mui/material';
import { useNavigate } from 'react-router-dom';
import VisibilityIcon from '@mui/icons-material/Visibility';
import { api } from '../services/api';
import { WebhookEvent } from '../types/webhook';
import { useNotification } from '../context/NotificationContext';

export const WebhookEventList: React.FC = () => {
  const [webhookEvents, setWebhookEvents] = useState<WebhookEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { showError } = useNotification();

  useEffect(() => {
    loadWebhookEvents();
  }, []);

  const loadWebhookEvents = async () => {
    try {
      setLoading(true);
      const data = await api.getAllWebhookEvents();
      setWebhookEvents(data);
    } catch (error) {
      showError('Failed to load webhook events');
      console.error('Failed to load webhook events:', error);
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

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          Webhook Events
        </Typography>
      </Box>

      {webhookEvents.length === 0 ? (
        <Alert severity="info">No webhook events found</Alert>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>JIRA Issue</TableCell>
                <TableCell>Project</TableCell>
                <TableCell>Summary</TableCell>
                <TableCell>Event Type</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Previous Status</TableCell>
                <TableCell>Received At</TableCell>
                <TableCell>Processed</TableCell>
                <TableCell align="center">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {webhookEvents.map((event) => (
                <TableRow key={event.id} hover>
                  <TableCell>
                    <Typography variant="body2" fontWeight="medium">
                      {event.jira_issue_key}
                    </Typography>
                  </TableCell>
                  <TableCell>{event.jira_project_key}</TableCell>
                  <TableCell>
                    <Typography variant="body2" noWrap sx={{ maxWidth: 250 }}>
                      {event.summary}
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={event.event_type}
                      color={getEventTypeColor(event.event_type)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{event.status || 'N/A'}</TableCell>
                  <TableCell>{event.previous_status || 'N/A'}</TableCell>
                  <TableCell>{formatDate(event.received_at)}</TableCell>
                  <TableCell>
                    {event.processed_at ? (
                      <Chip label="Yes" color="success" size="small" />
                    ) : (
                      <Chip label="No" color="default" size="small" />
                    )}
                  </TableCell>
                  <TableCell align="center">
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<VisibilityIcon />}
                      onClick={() => navigate(`/webhook-events/${event.id}`)}
                    >
                      View
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </Box>
  );
};
