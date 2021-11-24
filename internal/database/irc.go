package database

import (
	"context"
	"database/sql"

	"github.com/autobrr/autobrr/internal/domain"

	"github.com/rs/zerolog/log"
)

type IrcRepo struct {
	db *sql.DB
}

func NewIrcRepo(db *sql.DB) domain.IrcRepo {
	return &IrcRepo{db: db}
}

func (ir *IrcRepo) GetNetworkByID(id int64) (*domain.IrcNetwork, error) {

	row := ir.db.QueryRow("SELECT id, enabled, name, server, port, tls, pass, invite_command, nickserv_account, nickserv_password FROM irc_network WHERE id = ?", id)
	if err := row.Err(); err != nil {
		log.Fatal().Err(err)
		return nil, err
	}

	var n domain.IrcNetwork

	var pass, inviteCmd sql.NullString
	var nsAccount, nsPassword sql.NullString
	var tls sql.NullBool

	if err := row.Scan(&n.ID, &n.Enabled, &n.Name, &n.Server, &n.Port, &tls, &pass, &inviteCmd, &nsAccount, &nsPassword); err != nil {
		log.Fatal().Err(err)
	}

	n.TLS = tls.Bool
	n.Pass = pass.String
	n.InviteCommand = inviteCmd.String
	n.NickServ.Account = nsAccount.String
	n.NickServ.Password = nsPassword.String

	return &n, nil
}

func (ir *IrcRepo) DeleteNetwork(ctx context.Context, id int64) error {
	tx, err := ir.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `DELETE FROM irc_network WHERE id = ?`, id)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("error deleting network: %v", id)
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM irc_channel WHERE network_id = ?`, id)
	if err != nil {
		log.Error().Stack().Err(err).Msgf("error deleting channels for network: %v", id)
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Error().Stack().Err(err).Msgf("error deleting network: %v", id)
		return err

	}

	return nil
}

func (ir *IrcRepo) ListNetworks(ctx context.Context) ([]domain.IrcNetwork, error) {

	rows, err := ir.db.QueryContext(ctx, "SELECT id, enabled, name, server, port, tls, pass, invite_command, nickserv_account, nickserv_password FROM irc_network")
	if err != nil {
		log.Fatal().Err(err)
	}

	defer rows.Close()

	var networks []domain.IrcNetwork
	for rows.Next() {
		var net domain.IrcNetwork

		var pass, inviteCmd sql.NullString
		var tls sql.NullBool

		if err := rows.Scan(&net.ID, &net.Enabled, &net.Name, &net.Server, &net.Port, &tls, &pass, &inviteCmd, &net.NickServ.Account, &net.NickServ.Password); err != nil {
			log.Fatal().Err(err)
		}

		net.TLS = tls.Bool
		net.Pass = pass.String
		net.InviteCommand = inviteCmd.String

		networks = append(networks, net)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return networks, nil
}

func (ir *IrcRepo) ListChannels(networkID int64) ([]domain.IrcChannel, error) {

	rows, err := ir.db.Query("SELECT id, name, enabled FROM irc_channel WHERE network_id = ?", networkID)
	if err != nil {
		log.Fatal().Err(err)
	}
	defer rows.Close()

	var channels []domain.IrcChannel
	for rows.Next() {
		var ch domain.IrcChannel

		if err := rows.Scan(&ch.ID, &ch.Name, &ch.Enabled); err != nil {
			log.Fatal().Err(err)
		}

		channels = append(channels, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return channels, nil
}

func (ir *IrcRepo) StoreNetwork(network *domain.IrcNetwork) error {

	netName := toNullString(network.Name)
	pass := toNullString(network.Pass)
	inviteCmd := toNullString(network.InviteCommand)

	nsAccount := toNullString(network.NickServ.Account)
	nsPassword := toNullString(network.NickServ.Password)

	var err error
	if network.ID != 0 {
		// update record
		_, err = ir.db.Exec(`UPDATE irc_network
			SET enabled = ?,
			    name = ?,
			    server = ?,
			    port = ?,
			    tls = ?,
			    pass = ?,
			    invite_command = ?,
			    nickserv_account = ?,
			    nickserv_password = ?,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = ?`,
			network.Enabled,
			netName,
			network.Server,
			network.Port,
			network.TLS,
			pass,
			inviteCmd,
			nsAccount,
			nsPassword,
			network.ID,
		)
	} else {
		var res sql.Result

		res, err = ir.db.Exec(`INSERT INTO irc_network (
                         enabled,
                         name,
                         server,
                         port,
                         tls,
                         pass,
                         invite_command,
			    		 nickserv_account,
			             nickserv_password
                         ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			network.Enabled,
			netName,
			network.Server,
			network.Port,
			network.TLS,
			pass,
			inviteCmd,
			nsAccount,
			nsPassword,
		)
		if err != nil {
			log.Error().Stack().Err(err).Msg("error executing query")
			return err
		}

		network.ID, err = res.LastInsertId()
	}

	return err
}

func (ir *IrcRepo) StoreChannel(networkID int64, channel *domain.IrcChannel) error {
	pass := toNullString(channel.Password)

	var err error
	if channel.ID != 0 {
		// update record
		_, err = ir.db.Exec(`UPDATE irc_channel
			SET 
			    enabled = ?,
				detached = ?,
				name = ?,
				password = ?
			WHERE 
			      id = ?`,
			channel.Enabled,
			channel.Detached,
			channel.Name,
			pass,
			channel.ID,
		)
	} else {
		var res sql.Result

		res, err = ir.db.Exec(`INSERT INTO irc_channel (
                         enabled,
                         detached,
                         name,
                         password,
                         network_id
                         ) VALUES (?, ?, ?, ?, ?)`,
			channel.Enabled,
			true,
			channel.Name,
			pass,
			networkID,
		)
		if err != nil {
			log.Error().Stack().Err(err).Msg("error executing query")
			return err
		}

		channel.ID, err = res.LastInsertId()
	}

	return err
}