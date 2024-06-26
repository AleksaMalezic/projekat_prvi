package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Node struct {
	NodeID    int       `json:"id"`
	Name      string    `json:"name"`
	ParentID  *int      `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type NodeWithChildren struct {
	Node
	NodeChildren []NodeWithChildren `json:"children,omitempty"`
}

func getChildrenRecursive(db *sql.DB, parentID int) ([]NodeWithChildren, error) {
	rows, err := db.Query("SELECT id, name, parent_id, created_at FROM public.organisational_unit WHERE parent_id = $1", parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []NodeWithChildren
	for rows.Next() {
		var child Node
		var parentID sql.NullInt64
		if err := rows.Scan(&child.NodeID, &child.Name, &parentID, &child.CreatedAt); err != nil {
			return nil, err
		}
		if parentID.Valid {
			id := int(parentID.Int64)
			child.ParentID = &id
		}

		childWithChildren := NodeWithChildren{Node: child}
		childChildren, err := getChildrenRecursive(db, child.NodeID)
		if err != nil {
			return nil, err
		}
		childWithChildren.NodeChildren = childChildren

		children = append(children, childWithChildren)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return children, nil
}

func getNodeWithChildren(db *sql.DB, userID int) (NodeWithChildren, error) {
	var n Node
	var parentID sql.NullInt64

	err := db.QueryRow("SELECT id, name, parent_id, created_at FROM public.organisational_unit WHERE id=$1", userID).Scan(&n.NodeID, &n.Name, &parentID, &n.CreatedAt)
	if err != nil {
		return NodeWithChildren{}, err
	}
	if parentID.Valid {
		parentIDInt := int(parentID.Int64)
		n.ParentID = &parentIDInt
	}

	children, err := getChildrenRecursive(db, n.NodeID)
	if err != nil {
		return NodeWithChildren{}, err
	}

	return NodeWithChildren{Node: n, NodeChildren: children}, nil
}

func getNodeById(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	Node, err := getNodeWithChildren(db, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.IndentedJSON(http.StatusNotFound, gin.H{"message": "User not found"})
		} else {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.IndentedJSON(http.StatusOK, Node)
}

func postNode(c *gin.Context) {
	var newNode Node

	if err := c.BindJSON(&newNode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	var insertStmt string
	var err error
	if newNode.ParentID != nil {
		insertStmt = `INSERT INTO public.organisational_unit(name, parent_id, created_at) VALUES($1, $2, NOW()) RETURNING id, created_at`
		err = db.QueryRow(insertStmt, newNode.Name, *newNode.ParentID).Scan(&newNode.NodeID, &newNode.CreatedAt)
	} else {
		insertStmt = `INSERT INTO public.organisational_unit(name, created_at) VALUES($1, NOW()) RETURNING id, created_at`
		err = db.QueryRow(insertStmt, newNode.Name).Scan(&newNode.NodeID, &newNode.CreatedAt)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, newNode)
}

func putNode(c *gin.Context) {
	id := c.Param("id")
	var updatedNode Node

	if err := c.BindJSON(&updatedNode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	_, err := db.Exec("UPDATE public.organisational_unit SET name=$1, parent_id=$2 WHERE id=$3", updatedNode.Name, updatedNode.ParentID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func patchNode(c *gin.Context) {
	id := c.Param("id")
	var updatedNode Node

	if err := c.BindJSON(&updatedNode); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	query := "UPDATE public.organisational_unit SET "
	args := []interface{}{}
	argID := 1

	if updatedNode.Name != "" {
		query += fmt.Sprintf("name=$%d,", argID)
		args = append(args, updatedNode.Name)
		argID++
	}

	if updatedNode.ParentID != nil {
		query += fmt.Sprintf("parent_id=$%d,", argID)
		args = append(args, *updatedNode.ParentID)
		argID++
	}

	query = query[:len(query)-1]
	query += fmt.Sprintf(" WHERE id=$%d", argID)
	args = append(args, id)

	_, err := db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user updated successfully"})
}

func deleteNode(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM public.organisational_unit WHERE id=$1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func getNodes(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, parent_id, created_at FROM public.organisational_unit")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		var parentID sql.NullInt64
		if err := rows.Scan(&n.NodeID, &n.Name, &parentID, &n.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if parentID.Valid {
			id := int(parentID.Int64)
			n.ParentID = &id
		}
		nodes = append(nodes, n)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, nodes)
}
