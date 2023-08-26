1. Import the old workery database into your postgres server.

2. Run the following.

```bash
clear; go run main.go import_tenant;
clear; go run main.go import_user;
clear; go run main.go import_user_role;
clear; go run main.go change_password --email="bart@mikasoftware.com" --password="xxx";
clear; go run main.go import_insurance_requirement;
clear; go run main.go import_how_hear_about_us_item;
clear; go run main.go import_skill_set;
clear; go run main.go import_skill_set_insurance_requirement;
clear; go run main.go import_comment;
clear; go run main.go import_vehicle_type;
clear; go run main.go import_tag;
clear; go run main.go import_service_fee;
clear; go run main.go import_bulletins;
clear; go run main.go import_customer;
clear; go run main.go import_customer_comment;
clear; go run main.go import_customer_tag;
clear; go run main.go import_associate;
clear; go run main.go import_associate_vehicle_type;
clear; go run main.go import_associate_skill_set;
clear; go run main.go import_associate_comment;
clear; go run main.go import_associate_insurance_requirement;
clear; go run main.go import_associate_away_log;
clear; go run main.go import_associate_tag;
clear; go run main.go import_order;
clear; go run main.go import_activity_sheet;
clear; go run main.go import_order_comment;
clear; go run main.go import_order_skill_set;
clear; go run main.go import_order_tag;
clear; go run main.go import_order_invoice;
clear; go run main.go import_order_deposit;
clear; go run main.go import_task_item;
clear; go run main.go import_staff;
clear; go run main.go import_staff_comment;
clear; go run main.go import_attachment_download_to_tmp_dir;
clear; go run main.go import_attachment_upload_from_tmp_dir;
```
