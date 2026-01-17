import React from 'react';
import {
  AppBar,
  Box,
  Container,
  Drawer,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
} from '@mui/material';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import FolderIcon from '@mui/icons-material/Folder';
import CodeIcon from '@mui/icons-material/Code';
import WebhookIcon from '@mui/icons-material/Webhook';

const drawerWidth = 240;

export const Layout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    { text: 'Projects', icon: <FolderIcon />, path: '/projects' },
    { text: 'Developments', icon: <CodeIcon />, path: '/developments' },
    { text: 'Webhook Events', icon: <WebhookIcon />, path: '/webhook-events' },
  ];

  return (
    <Box sx={{ display: 'flex' }}>
      <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            SDLC Agents - Backoffice
          </Typography>
        </Toolbar>
      </AppBar>
      <Drawer
        variant="permanent"
        sx={{
          width: drawerWidth,
          flexShrink: 0,
          '& .MuiDrawer-paper': {
            width: drawerWidth,
            boxSizing: 'border-box',
          },
        }}
      >
        <Toolbar />
        <Box sx={{ overflow: 'auto' }}>
          <List>
            {menuItems.map((item) => (
              <ListItem key={item.text} disablePadding>
                <ListItemButton
                  selected={location.pathname.startsWith(item.path)}
                  onClick={() => navigate(item.path)}
                >
                  <ListItemIcon>{item.icon}</ListItemIcon>
                  <ListItemText primary={item.text} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Box>
      </Drawer>
      <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
        <Toolbar />
        <Container maxWidth="xl">
          <Outlet />
        </Container>
      </Box>
    </Box>
  );
};
