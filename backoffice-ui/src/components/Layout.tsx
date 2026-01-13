import React from 'react';
import {
  AppBar,
  Box,
  Container,
  Toolbar,
  Typography,
} from '@mui/material';
import { Outlet } from 'react-router-dom';

export const Layout: React.FC = () => {
  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100vh' }}>
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            SDLC Agents - Backoffice
          </Typography>
        </Toolbar>
      </AppBar>
      <Container maxWidth="xl" sx={{ mt: 4, mb: 4, flex: 1 }}>
        <Outlet />
      </Container>
    </Box>
  );
};
