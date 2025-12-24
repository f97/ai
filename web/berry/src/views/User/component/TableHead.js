import { TableCell, TableHead, TableRow } from '@mui/material';

const UsersTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>ID</TableCell>
        <TableCell>Username</TableCell>
        <TableCell>分组</TableCell>
        <TableCell>统计信息</TableCell>
        <TableCell>用户角色</TableCell>
        <TableCell>绑定</TableCell>
        <TableCell>Status</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default UsersTableHead;
