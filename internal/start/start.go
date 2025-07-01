package start

import "gitlab.com/dgb9/todo-api/internal/controller"

func Start() error {
	// load the data configuration
	config, err := loadConfig()
	if err != nil {
		return err
	}

	// connect to database and then test
	db, err := connectDb(config.Db)
	if err != nil {
		return err
	}
	// configure the router and start
	err = controller.StartRouter(db, config.Server)
	if err != nil {
		return err
	}

	return nil
}
