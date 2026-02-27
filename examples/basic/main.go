// This is an example showing how to use the Shopline SDK.
//
// Usage:
//
//	export SHOPLINE_APP_KEY="your-app-key"
//	export SHOPLINE_APP_SECRET="your-app-secret"
//	export SHOPLINE_HANDLE="your-store-handle"
//	export SHOPLINE_TOKEN="your-access-token"
//	go run examples/basic/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	shopline "github.com/imokyou/slshop"
	"github.com/imokyou/slshop/customer"
	"github.com/imokyou/slshop/order"
)

func main() {
	app := shopline.App{
		AppKey:    os.Getenv("SHOPLINE_APP_KEY"),
		AppSecret: os.Getenv("SHOPLINE_APP_SECRET"),
	}

	handle := os.Getenv("SHOPLINE_HANDLE")
	token := os.Getenv("SHOPLINE_TOKEN")

	if handle == "" || token == "" {
		log.Fatal("Please set SHOPLINE_HANDLE and SHOPLINE_TOKEN environment variables")
	}

	// Create client with retry and specific API version
	client, err := shopline.NewClient(app, handle, token,
		shopline.WithVersion(shopline.APIVersion20251201),
		shopline.WithRetry(3),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// === Get Shop Info ===
	fmt.Println("=== Shop Info ===")
	shop, err := client.Store.GetShop(ctx)
	if err != nil {
		log.Printf("Failed to get shop: %v", err)
	} else {
		fmt.Printf("Shop: %s (%s)\n", shop.Name, shop.Domain)
	}

	// === List Products ===
	fmt.Println("\n=== Products ===")
	products, err := client.Product.List(ctx, nil)
	if err != nil {
		log.Printf("Failed to list products: %v", err)
	} else {
		fmt.Printf("Found %d products:\n", len(products))
		for _, p := range products {
			fmt.Printf("  - [%d] %s (variants: %d)\n", p.ID, p.Title, len(p.Variants))
		}
	}

	// === Count Products ===
	count, err := client.Product.Count(ctx, nil)
	if err != nil {
		log.Printf("Failed to count products: %v", err)
	} else {
		fmt.Printf("Total products: %d\n", count)
	}

	// === List Orders ===
	fmt.Println("\n=== Orders ===")
	orders, err := client.Order.List(ctx, &order.ListOptions{
		Status: "any",
	})
	if err != nil {
		log.Printf("Failed to list orders: %v", err)
	} else {
		fmt.Printf("Found %d orders:\n", len(orders))
		for _, o := range orders {
			fmt.Printf("  - %s: %s %s (status: %s)\n", o.Name, o.Currency, o.TotalPrice, o.FinancialStatus)
		}
	}

	// === List Customers ===
	fmt.Println("\n=== Customers ===")
	customers, err := client.Customer.List(ctx, &customer.ListOptions{})
	if err != nil {
		log.Printf("Failed to list customers: %v", err)
	} else {
		fmt.Printf("Found %d customers:\n", len(customers))
		for _, c := range customers {
			fmt.Printf("  - [%d] %s %s (%s)\n", c.ID, c.FirstName, c.LastName, c.Email)
		}
	}

	fmt.Println("\nDone!")
}
