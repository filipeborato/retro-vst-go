package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"retro-vst-go/domain"
	"retro-vst-go/repository"
)

type ProductInput struct {
	ProductName string  `json:"product_name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gte=0"`
}

func GetProductsHandler(productRepo repository.ProductRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		products, err := productRepo.GetAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

func GetProductByIDHandler(productRepo repository.ProductRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		product, err := productRepo.GetByID(uint(id))
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

func CreateProductHandler(productRepo repository.ProductRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input ProductInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		product := domain.Product{
			ProductName: input.ProductName,
			Description: input.Description,
			Price:       input.Price,
		}

		if err := productRepo.Create(&product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		c.JSON(http.StatusCreated, product)
	}
}

func UpdateProductHandler(productRepo repository.ProductRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var input ProductInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		product, err := productRepo.GetByID(uint(id))
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}
			return
		}

		product.ProductName = input.ProductName
		product.Description = input.Description
		product.Price = input.Price

		if err := productRepo.Update(product); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

func DeleteProductHandler(productRepo repository.ProductRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		_, err = productRepo.GetByID(uint(id))
		if err != nil {
			if errors.Is(err, repository.ErrProductNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product"})
			}
			return
		}

		if err := productRepo.Delete(uint(id)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}
