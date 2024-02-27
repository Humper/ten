import { DateInput, Edit, NumberInput, SimpleForm, TextInput } from 'react-admin';

export const UserEdit = () => (
    <Edit>
        <SimpleForm>

            <TextInput source="Name" />
            <TextInput source="Email" />
            <TextInput source="Role" />
            <TextInput source="AllowedIPs" />
        </SimpleForm>
    </Edit>
);