import { ArrayField, ChipField, Datagrid, DateField, List, NumberField, SingleFieldList, TextField } from 'react-admin';

export const UserList = () => (
    <List>
        <Datagrid rowClick="edit" bulkActionButtons={false}>
            <TextField source="Name" />
            <TextField source="Email" />
            <TextField source="Role" />
        </Datagrid>
    </List>
);