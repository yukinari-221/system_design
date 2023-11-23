/* placeholder file for JavaScript */

const confirm_delete = (id) => {
    if(window.confirm(`Task ${id} を削除します．よろしいですか？`)) {
        location.href = `/task/delete/${id}`;
    }
}

const confirm_update = (id) => {
    if(window.confirm(`Task ${id} を更新します．よろしいですか？`)) {
        return ture;
    }
    return false;
}

const confirm_deleteuser = (name) => {
    if(window.confirm(`user ${name} を削除します．よろしいですか？`)) {
        return ture;
    }
    return false;
}
