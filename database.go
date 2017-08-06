package main

func databaseinit() {
	tx, err := lb.db.Begin()
        if err == nil {
                _, err = tx.Exec(`CREATE SEQUENCE accessorys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MAXVALUE
    NO MINVALUE
    CACHE 1;`)
        }
	if err == nil {
                _, err = tx.Exec(`CREATE TABLE accessorys (id integer NOT NULL, key string, name string, serial_number string, manu_facturer string, model string, pin string, accessory_type string, created_at timestamp without time zone,updated_at timestamp without time zone)`)
        }
        if err == nil {
                _, err = tx.Exec(`CREATE INDEX index_accessorys_on_id on accessorys USING btree (id);`)
        }
        if err == nil {
                _, err = tx.Exec(`CREATE UNIQUE INDEX index_accessorys_on_key on accessorys USING btree (key);`)
        }
        if err != nil {
                tx.Rollback()
                return err
        }
        if err = tx.Commit(); err != nil {
                tx.Rollback()
                return err
        }
        return err

}
