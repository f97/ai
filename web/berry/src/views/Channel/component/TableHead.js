import { TableCell, TableHead, TableRow } from '@mui/material';

const ChannelTableHead = () => {
  return (
    <TableHead>
      <TableRow>
        <TableCell>ID</TableCell>
        <TableCell>Name</TableCell>
        <TableCell>分组</TableCell>
        <TableCell>Type</TableCell>
        <TableCell>Status</TableCell>
        <TableCell>响应时间</TableCell>
        <TableCell>已消耗</TableCell>
        <TableCell>Balance</TableCell>
        <TableCell>优先级</TableCell>
        <TableCell>操作</TableCell>
      </TableRow>
    </TableHead>
  );
};

export default ChannelTableHead;
