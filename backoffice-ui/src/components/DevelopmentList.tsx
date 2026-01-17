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
import { Development } from '../types/development';
import { useNotification } from '../context/NotificationContext';

export const DevelopmentList: React.FC = () => {
  const [developments, setDevelopments] = useState<Development[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { showError } = useNotification();

  useEffect(() => {
    loadDevelopments();
  }, []);

  const loadDevelopments = async () => {
    try {
      setLoading(true);
      const data = await api.getAllDevelopments();
      setDevelopments(data);
    } catch (error) {
      showError('Failed to load developments');
      console.error('Failed to load developments:', error);
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

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4" component="h1">
          Developments
        </Typography>
      </Box>

      {developments.length === 0 ? (
        <Alert severity="info">No developments found</Alert>
      ) : (
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>JIRA Issue</TableCell>
                <TableCell>Project</TableCell>
                <TableCell>Repository</TableCell>
                <TableCell>Branch</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>Created At</TableCell>
                <TableCell>PR/MR</TableCell>
                <TableCell align="center">Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {developments.map((dev) => (
                <TableRow key={dev.id} hover>
                  <TableCell>
                    <Typography variant="body2" fontWeight="medium">
                      {dev.jira_issue_key}
                    </Typography>
                  </TableCell>
                  <TableCell>{dev.jira_project_key}</TableCell>
                  <TableCell>
                    <Typography variant="body2" noWrap sx={{ maxWidth: 200 }}>
                      {dev.repository_url}
                    </Typography>
                  </TableCell>
                  <TableCell>{dev.branch_name}</TableCell>
                  <TableCell>
                    <Chip
                      label={dev.status}
                      color={getStatusColor(dev.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{formatDate(dev.created_at)}</TableCell>
                  <TableCell>
                    {dev.pr_mr_url ? (
                      <Button
                        size="small"
                        href={dev.pr_mr_url}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        View PR
                      </Button>
                    ) : (
                      <Typography variant="body2" color="text.secondary">
                        N/A
                      </Typography>
                    )}
                  </TableCell>
                  <TableCell align="center">
                    <Button
                      variant="outlined"
                      size="small"
                      startIcon={<VisibilityIcon />}
                      onClick={() => navigate(`/developments/${dev.id}`)}
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
